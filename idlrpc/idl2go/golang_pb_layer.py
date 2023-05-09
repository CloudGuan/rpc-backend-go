import json
import sys
import os


def json_has_attr(json_obj, attr_name):
    if json_obj is None:
        return False

    if attr_name is None:
        return False

    return (attr_name in json_obj) and (json_obj[attr_name] is not None)

class FileGenerator:
    def __init__(self, file_name, pb_json = None):
        self.file_name = file_name
        self.file_name_split = file_name.split(".")
        self.pb_json = pb_json
        self.gen_struct_pb_file()

    def gen_struct_pb_file(self):
        file = open(self.file_name_split[0] + ".service.proto", "w")
        file.write("// Machine generated code\n\n")
        file.write('syntax = "proto3";\n\n')
        file.write("package " + self.file_name_split[0] + ";\n\n")
        file.write('option go_package="idldata/pbdata";\n\n')
        file.write('message ServiceBoxTraceInfo {\n')
        file.write('\tstring tracer_id = 1;\n\tuint64 span_id = 2;\n\tuint64 parent_span_id = 3;\n\tstring user_data = 4;\n}\n\n')
        file.write('message ExceptionInfo {\n')
        file.write('\tstring name = 1;\n\tstring detail = 2;\n}\n\n')
        file.write('')
        file.write('message KeyValue {\n')
        file.write('    string key = 1;\n')
        file.write('    string value = 2;\n')
        file.write('}\n')
        file.write('message Context {\n')
        file.write('    repeated KeyValue info = 1;\n')
        file.write('}\n')

        self.gen_enum_pb_file(file)
        if json_has_attr(self.pb_json, "structs"):
            for struct in self.pb_json["structs"]:
                file.write("message " + struct["name"] + " {\n")
                n = 1
                if json_has_attr(struct, "fields"):
                    for field in struct["fields"]:
                        if field["IdlType"] == "dict":
                            file.write("    " + field["type"] + "<" + field["key"]["type"] + "," +   field["value"]["type"] + "> " + field["name"] + " = " + str(n) + ";\n")
                        elif field["IdlType"] == "seq" or field["IdlType"] == "set":
                            file.write("    " + field["type"] + " " + field["key"]["type"] + " " + field["name"] + " = " + str(n) + ";\n")
                        else:
                            file.write("    " + field["type"] + " " + field["name"] + " = " + str(n) + ";\n")
                        n += 1
                file.write("}\n\n")
        self.gen_method_pb_file(file)
        self.gen_method_ret_pb(file)
        file.close()


    def gen_method_ret_pb(self, file):
        for service in self.pb_json["services"]:
            for method in service["methods"]:
                file.write("message " + service["name"] + "_" + method["name"] + "_ret {\n")
                if method["retType"]["IdlType"] == "dict":
                    file.write("    " + method["retType"]["type"] + "<" + method["retType"]["key"]["type"] + "," +   method["retType"]["value"]["type"] + "> " + " ret1 = 1;\n")
                elif method["retType"]["IdlType"] == "seq" or method["retType"]["IdlType"] == "set":
                    file.write("    " + method["retType"]["type"] + " " + method["retType"]["key"]["type"] + " ret1 = 1;\n")
                elif method["retType"]["IdlType"] == "void":
                    pass
                else:
                    file.write("    " + method["retType"]["type"] + " ret1 = 1;\n")
                #add trace info and exception info 
                file.write("    ServiceBoxTraceInfo trace_info = 2;\n")
                file.write("    ExceptionInfo exec_info = 3;\n")
                file.write("    Context ctx = 4;\n")
                file.write("}\n\n")

    

    def gen_enum_pb_file(self,  file):
        if not json_has_attr(self.pb_json, "enums"):
            return

        for enum in self.pb_json["enums"]:
            file.write('enum {ename} {{\n    default_{ename}=0;\n'.format(ename=enum["name"]));
            if "fields" in enum:
                for filed in enum["fields"]:
                    file.write('    {fname}={fvalue};\n'.format(fname=filed["name"], fvalue=filed["value"]))
                file.write('}\n\n')


    def gen_method_pb_file(self, file):
        for service in self.pb_json["services"]:
            for method in service["methods"]:
                file.write("message " + service["name"] + "_" + method["name"] + "_args {\n")
                n = 1
                for arg in method["arguments"]:
                    if arg["IdlType"] == "dict":
                        file.write("    " + arg["type"] + "<" + arg["key"]["type"] + "," +   arg["value"]["type"] + ">")
                    elif arg["IdlType"] == "seq" or arg["IdlType"] == "set":
                        file.write("    " + arg["type"] + " " + arg["key"]["type"])
                    elif arg["IdlType"] == "void":
                        continue
                    else:
                        file.write("    " + arg["type"])
                    file.write(" arg" + str(n) + " = " + str(n) + ";\n") 
                    n += 1
                # add extern config 
                file.write("    ServiceBoxTraceInfo trace_info = "+str(n) +";\n")
                file.write("    Context ctx = "+str(n+1) +";\n")
                file.write("}\n\n")


if __name__ == "__main__":
    pbjson = json.load(open(sys.argv[2]))
    FileGenerator(sys.argv[1], pbjson)