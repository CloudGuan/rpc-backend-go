package idlrpc

import (
	"fmt"
	"github.com/CloudGuan/rpc-backend-go/idlrpc/internal/logger"
	"github.com/CloudGuan/rpc-backend-go/idlrpc/pkg/errors"
	"github.com/CloudGuan/rpc-backend-go/idlrpc/pkg/protocol"
	"github.com/CloudGuan/rpc-backend-go/idlrpc/pkg/transport"
	"sync"
	"sync/atomic"
)

type ProxyMap map[ProxyId]IProxy   //key, proxyid, value proxy interface
type Trans2Proxy map[uint64]IProxy //service uuid, proxy interface

// tpWrapper transport proxy wrapper
// one transport has mutil service uuid proxy
type tpWrapper struct {
	transId  uint32      //transportID
	proxyMap Trans2Proxy //service to proxy cache
}

type tpCache map[uint32]*tpWrapper                          //transport id, tp wrapper
type outSideTpCache map[protocol.GlobalIndexType]*tpWrapper //outside transport cache

// ProxyManager  manager connect proxy for rpc framework, multiple may be read & write
// TODO add transport id 2 service uid cache
type ProxyManager struct {
	pId               ProxyId         //proxy instance id
	proxyMap          ProxyMap        //proxy instance cache
	proxyCache        tpCache         //transport to proxy cache
	outsideProxyCache outSideTpCache  //outside transport cache
	factory           proxyFactoryMap //proxy factory
	mux               sync.RWMutex    //mutex
}

// newProxyManager create new proxy manager, init proxymap and mutex
func newProxyManager() *ProxyManager {
	return &ProxyManager{
		pId:               1,
		proxyMap:          make(ProxyMap),
		proxyCache:        make(tpCache),
		outsideProxyCache: make(outSideTpCache),
		factory:           make(proxyFactoryMap),
		mux:               sync.RWMutex{},
	}
}

func (p *ProxyManager) GeneProxyId() ProxyId {
	return ProxyId(atomic.AddUint32((*uint32)(&p.pId), 1))
}

func (p *ProxyManager) Add(proxy IProxy) error {
	if proxy == nil {
		logger.Error("[ProxyManager] Invalid invalid proxy interface")
		return errors.NewRpcError(errors.CommErr, "invalid proxy interface")
	}

	if p == nil {
		logger.Error("[ProxyManager] proxy manager not init while add proxy: %d, %s", proxy.GetUUID(), proxy.GetSrvName())
		return errors.NewRpcError(errors.CommErr, "invalid proxy manager")
	}

	//generate proxy id
	proxyId := p.GeneProxyId()
	proxy.SetID(proxyId)

	trans := proxy.GetTransport()
	if trans == nil || trans.IsClose() {
		return errors.ErrTransClose
	}

	//write lock
	p.mux.Lock()
	defer p.mux.Unlock()

	if _, ok := p.proxyMap[proxyId]; ok {
		logger.Error("[IProxy] %s,%d,%d proxy has been exist", proxy.GetSrvName(), proxy.GetUUID(), proxy.GetID())
		return errors.ErrProxyInvalid
	}
	//proxy id has been set while create thie struct
	p.proxyMap[proxyId] = proxy

	//add to trans port cache

	var tp *tpWrapper
	var ok bool
	// 根据 globalIndex 是否为零，判断从哪里获取缓存
	if proxy.GetGlobalIndex() != InvalidGlobalIndex {
		tp, ok = p.outsideProxyCache[proxy.GetGlobalIndex()]
	} else {
		tp, ok = p.proxyCache[trans.GetID()]
	}

	if !ok {
		tp = &tpWrapper{
			transId:  trans.GetID(),
			proxyMap: make(Trans2Proxy),
		}

		//p.proxyCache[tp.transId] = tp
		if proxy.GetGlobalIndex() != InvalidGlobalIndex {
			p.outsideProxyCache[proxy.GetGlobalIndex()] = tp
		} else {
			p.proxyCache[trans.GetID()] = tp
		}
	}
	tp.proxyMap[proxy.GetUUID()] = proxy

	return nil
}

func (p *ProxyManager) Get(proxyId ProxyId) (IProxy, error) {
	if p == nil {
		logger.Error("[ProxyManager] proxy manager init error while get proxyId %d", proxyId)
		return nil, errors.NewRpcError(errors.CommErr, "invalid proxy manager")
	}

	p.mux.RLock()
	defer p.mux.RUnlock()

	proxy, ok := p.proxyMap[proxyId]
	if !ok {
		return nil, errors.NewProxyNotExit(uint32(proxyId))
	}

	return proxy, nil
}

func (p *ProxyManager) Destroy(proxyId ProxyId) error {
	if p == nil {
		logger.Error("[ProxyManager] %d,0,0 proxy manager is valid", proxyId)
		return errors.NewRpcError(errors.CommErr, "invalid proxy manager")
	}

	//lock map
	p.mux.Lock()
	defer p.mux.Unlock()

	proxy, ok := p.proxyMap[proxyId]
	if !ok {
		return nil
	}

	delete(p.proxyMap, proxyId)

	//delete from trans
	transId := proxy.GetTransport().GetID()
	if tp, ok := p.proxyCache[transId]; ok {
		delete(tp.proxyMap, proxy.GetUUID())
	}

	return nil
}

func (p *ProxyManager) addCreator(uuid uint64, creator ProxyCreator) {
	p.mux.Lock()
	defer p.mux.Unlock()
	p.factory[uuid] = creator
}

func (p *ProxyManager) getOrCreateProxy(uuid uint64, globalIndex protocol.GlobalIndexType, trans transport.ITransport) IProxy {
	p.mux.Lock()
	defer p.mux.Unlock()

	creator, ok := p.factory[uuid]
	if !ok {
		return nil
	}

	//tp, ok = p.proxyCache[trans.GetID()]
	var tp *tpWrapper
	// 根据 globalIndex 是否为零，判断从哪里获取缓存
	if globalIndex != InvalidGlobalIndex {
		tp, ok = p.outsideProxyCache[globalIndex]
	} else {
		tp, ok = p.proxyCache[trans.GetID()]
	}

	if !ok {
		tp = &tpWrapper{
			transId:  trans.GetID(),
			proxyMap: make(Trans2Proxy),
		}

		if globalIndex != InvalidGlobalIndex {
			p.outsideProxyCache[globalIndex] = tp
		} else {
			p.proxyCache[tp.transId] = tp
		}
	}

	proxy, ok := tp.proxyMap[uuid]
	if !ok || !proxy.IsConnected() {
		// not cache proxy
		proxy = creator(trans)
		// create error
		if proxy == nil {
			return nil
		}
		proxy.SetID(p.GeneProxyId())
		proxy.SetGlobalIndex(globalIndex)
		p.proxyMap[proxy.GetID()] = proxy
		tp.proxyMap[uuid] = proxy
	}
	return proxy
}

func (p *ProxyManager) closeOutsideProxy(outsideId protocol.GlobalIndexType) error {
	if outsideId == InvalidGlobalIndex {
		logger.Warn("Failed to clear expired connections because the parameter is invalid")
		return fmt.Errorf("Failed to clear expired connections because the parameter is invalid! ")
	}

	// 上锁, 并且直接清理掉对应缓存
	var tp *tpWrapper
	var ok bool
	p.mux.Lock()
	if tp, ok = p.outsideProxyCache[outsideId]; ok {
		// 如果存在， 从缓存中删除
		delete(p.outsideProxyCache, outsideId)
	}
	p.mux.Unlock()

	if tp == nil {
		return nil
	}

	// 设置proxy 关闭状态，保证外部缓存到的proxy能够探知到proxy不可用
	for _, p := range tp.proxyMap {
		p.close()
	}

	// 清理残留proxy
	tp.proxyMap = nil

	return nil
}
