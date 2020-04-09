package parse 

import (
        "strings"
        "regexp"
        "fmt"
)


func Findpart(par string,msg string) string {
     stop_idx := 0
     brace_counter := 0
     r, _ := regexp.Compile("{|}")
     start_idx:= strings.Index(msg,par)
     if start_idx == -1 {
        return ""
     }
     start_idx=start_idx+len(par)
     search_str := msg[start_idx:]
     for _,idx:=range r.FindAllStringIndex(search_str,-1){
         stop_idx =idx[0]
         defer func(){
           if err:=recover();err!=nil{
               fmt.Println(" error stop_idx ",stop_idx,len(search_str))
               }
            }()
         if string(search_str[stop_idx]) == "{" {
            brace_counter = brace_counter + 1
         }else{
            brace_counter = brace_counter - 1
         }
         if brace_counter == 0 {
            break
         }
     }
     search_str = strings.TrimSpace(search_str[:stop_idx + 1])
     return search_str
}






func Change_type(msg string) map[string]string {
     par:=`(?P<pre>mongod\.\d+\[\d+\]\: \[conn\d+\])\s+(?P<op>\w+)\s+(?P<dbtb>\S+)\s+(?P<command>\S+)`
     result:=Regexp_map(par,msg) 
     //for k,v:=range result {
     //    fmt.Println(k,v)
     //}
     
     var rst map[string]string
     if len(result)>0 {
        switch result["op"] {
               case "command":
                    if ! strings.HasPrefix(result["dbtb"],`local.oplog`) && ! strings.HasPrefix(result["dbtb"],`config.chunks`){
                       rst=command_type_get(msg)
                       dtlst:=strings.Split(result["dbtb"],".")
                       if len(dtlst)>1 && ! strings.HasPrefix(dtlst[1],`$`){
                          rst["table"]=strings.Split(result["dbtb"],".")[1] 
                       }
                       rst["db"]=strings.Split(result["dbtb"],".")[0]
                       
                   }

               case "build":
                    rst=make(map[string]string)
                    rst["optp"]="build Index"
                    rst["msg"]=msg

               default:
                    rst=make(map[string]string)

                    var r map[string]string
                    switch result["op"] {
                        case "update":
                             rst["optp"]="update"
                             r=Regexp_map(`(?P<key> docsExamined):(?P<value>\w+)`,msg)
                             if v,ok:=r["value"];ok{rst["doc_scan"]=v}
                             r=Regexp_map(`(?P<pre>] update )(?P<db>\w+).(?P<tab>\w+)(?P<op> query)`,msg)  // update english.report query:

                        case "remove":
                             rst["optp"]="remove"
                             r=Regexp_map(`(?P<key> ndeleted):(?P<value>\w+)`,msg)
                             if v,ok:=r["value"];ok{rst["delete_num"]=v}
                             r=Regexp_map(`(?P<pre>] remove )(?P<db>\w+).(?P<tab>\w+)(?P<op> query)`,msg)  // remove english.report query:

                        case "query":
                             rst["optp"]="query"
                             r=Regexp_map(`(?P<key> nreturned): (?P<value>\w+),`,msg)
                             if v,ok:=r["value"];ok{rst["return_rows"]=v}
                             r=Regexp_map(`(?P<key> docsExamined):(?P<value>\w+)`,msg)
                             if v,ok:=r["value"];ok{rst["doc_scan"]=v}
                             rst["plan"]=Findpart("planSummary: ",msg)
                             if strings.HasPrefix(rst["plan"],`COLLSCAN`){
                                rst["plan"]="COLLSCAN"
                             }
                             if strings.HasPrefix(rst["plan"],`IDHACK`){
                                rst["plan"]="IDHACK"
                             }
                                  
                             r=Regexp_map(`(?P<pre>] query )(?P<db>\w+).(?P<tab>\w+)(?P<op> query)`,msg)   // query english.report query:
                             if len(r)==0{
                                r=Regexp_map(`(?P<pre>] query )(?P<db>\w+).(?P<tab>\w+)(?P<op> planSummary)`,msg)
                             } 
                        case "getmore":
                             dtlst:=strings.Split(result["dbtb"],".")
                             if ! strings.HasPrefix(dtlst[1],`$`){
                                rst["table"]=strings.Split(result["dbtb"],".")[1]
                             }
                             rst["db"]=strings.Split(result["dbtb"],".")[0]
                             rst["optp"]="getmore" 
                        default:
                             m:=Regexp_map(`(?P<pre>\s\[conn\d+\]\s)(?P<msg_head>.+)(?P<cmd>command:)\s+(?P<type>[^\{]+)(?P<typend>{)(?P<msg>.+)`,msg)
                             qpart:=Findpart(Get_v(m,"type"),msg)
                             
                             per,_:=regexp.Compile(`(?i) error | exception:| error:`)
                             var last_msg string
                             if lmsg:=Get_v(m,"msg");len(lmsg)>len(qpart){
                                last_msg=lmsg[len(qpart):len(lmsg)]
                             }
                             if per.MatchString(Get_v(m,"msg_head")) || per.MatchString(last_msg){
                                rst["optp"]="exception"
                             }
                             if len(qpart)>0{
                                rst["query_part"]=last_msg
                             }
                             rst["msg"]=msg
                             

                             
                              
                    }
                    if v,ok:=r["db"];ok{
                     rst["db"]=v
                     rst["query_part"]=Findpart("query: ",msg)
                    } 
                    if v,ok:=r["tab"];ok{
                     rst["table"]=v
                    } 
                    
        }
     }else{
       par=`(?P<pre>mongod\.\d+\[\d+\]\: \[conn\d+\])\s+CMD\: .+`
       if ok,_:=regexp.MatchString(par, msg);ok{
          rst=make(map[string]string)
          rst["optp"]="ddl"   
          rst["msg"]=msg
       } 
     }
     return rst
}




func command_type_get(msg string) map[string]string {
     rmap:=make(map[string]string)
     par:=`\s+command\:\s+(?P<tp>\w+)`
     result:=Regexp_map(par,msg)
     tp:=result["tp"]
     if tp=="" {return rmap}
     switch tp {
         case "update":
               query:=Findpart(`[ { q:`,Findpart("updates:",msg))
               rmap["optp"]="update"
               tmp:=Regexp_map(`(?P<pre>update:)\s(?P<tab>\S+)`,msg)
               if v,ok:=tmp["tab"];ok{
                  rmap["table"]=v
                  rmap["query_part"]=query
               }
         case "find":
               rmap["query_part"]=Findpart("filter: ",msg)
               rmap["optp"]="find"
               rmap["sort"]=Findpart("sort: ",msg)
               rmap["plan"]=Findpart("planSummary: ",msg)
               if strings.HasPrefix(rmap["plan"],`COLLSCAN`){
                  rmap["plan"]="COLLSCAN"
               }

               r:=Regexp_map(`(?P<key> docsExamined):(?P<value>\w+)`,msg)
               if v,ok:=r["value"];ok{rmap["doc_scan"]=v}

               r=Regexp_map(`(?P<key> limit): (?P<value>\w+),`,msg)
               if v,ok:=r["value"];ok{rmap["limit"]=v}

               r=Regexp_map(`(?P<key> find): (?P<value>\w+),`,msg)
               if v,ok:=r["value"];ok{rmap["table"]=v}


               r=Regexp_map(`(?P<key> nreturned): (?P<value>\w+),`,msg)
               if v,ok:=r["value"];ok{rmap["return_rows"]=v}
         case "count":
               rmap["optp"]="count"
               rmap["query_part"]=Findpart("query: ",msg)
               rmap["plan"]=Findpart("planSummary: ",msg)

               if strings.HasPrefix(rmap["plan"],`COLLSCAN`){
                  rmap["plan"]="COLLSCAN"
               }

               r:=Regexp_map(`(?P<key> count): (?P<value>\w+),`,msg)
               if v,ok:=r["value"];ok{rmap["table"]=v}
         case "findAndModify":
               rmap["optp"]="findandmodify"
               rmap["query_part"]=Findpart("query: ",msg)

               r:=Regexp_map(`(?P<key> docsExamined):(?P<value>\w+)`,msg)
               if v,ok:=r["value"];ok{rmap["doc_scan"]=v}

               r=Regexp_map(`(?P<key> findandmodify): (?P<value>\S+),`,msg)
               if v,ok:=r["value"];ok{rmap["table"]=v}
         default:
                m:=Regexp_map(`(?P<pre>\s\[conn\d+\]\s)(?P<msg_head>.+)(?P<cmd>command:)\s+(?P<type>[^\{]+)(?P<typend>{)(?P<msg>.+)`,msg)
                qpart:=Findpart(Get_v(m,"type"),msg)
                rhd,_:=regexp.Compile(`\S+:`)
                if hop:=rhd.FindAllString(qpart,1);qpart!="" && len(hop)>0 && strings.TrimRight(hop[0],":")==strings.TrimSpace(Get_v(m,"type")) {
                   if tab,err:=Regexp_map(`(?P<key>`+hop[0]+`)\s+(?P<tab>\S+),`,qpart)["tab"];err{
                      tab=strings.Trim(tab,`"`)
                      rmap["optp"]=strings.TrimRight(hop[0],":")
                      rmap["table"]=tab
                   }
                }
                per,_:=regexp.Compile(`(?i) error | exception:| error:`)
                var last_msg string
                if lmsg:=Get_v(m,"msg");len(lmsg)>len(qpart){
                   last_msg=lmsg[len(qpart):len(lmsg)]
                }
                if per.MatchString(Get_v(m,"msg_head")) || per.MatchString(last_msg){
                   rmap["optp"]="exception" 
                }
                if len(qpart)>0{
                   rmap["query_part"]=qpart
                }
                
     }
     return rmap
}






























