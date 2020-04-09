package parse

import (
       "fmt"
       "crypto/md5"
       "time"
       "io"
       "regexp"
       "strconv"
)



func (sqlite *DB_save)Format_output(query map[string]string,line string,account int,ltm time.Time) (string ,string) {
     now:=ltm.Format("2006-01-02 15:04:05")
     optp:=Get_v(query,"optp")
     etime,_:=strconv.Atoi(Get_v(query,"ms"))
     sqlid:=gen_qid(query)
     hst:=Get_v(query,"host")
     mp:=&Save_alert_st{
     qtm:int(ltm.Unix()),
     atm:int(time.Now().Unix()),
     etime:etime,
     qid:sqlid,
     db:Get_v(query,"db"),
     table:Get_v(query,"table"),
     host:hst,
     optp:optp,
     qmessage:Regexp_map(`(?P<pre>\s\[conn\d+\]\s)(?P<msg>.+)`,line)["msg"]}
     sqlite.Save_alert(mp) 
     //sqlite.Save_ack(sqlid,int(time.Now().Unix()))
     if _,tm,_,_:=sqlite.Find_ack(sqlid);tm>0 {
        if sqlite.Debug{fmt.Println("**skip ack sqlid ",sqlid," : ",mp.qmessage)}
        return "",""
     }
     if sqlite.Find_alert_count(sqlid,account){
        if _,tm,_,_:=sqlite.Find_ack(sqlid);tm==0 { 
           amsg:=fmt.Sprintf("auto ack after %d times",account)
           if sqlite.Debug{fmt.Println("**sqlid",sqlid," auto ack after ",account," times ")}
           sqlite.Save_ack(sqlid,int(time.Now().Unix()),"","",amsg)
        }
        return "",""
     }
     return sqlid,fmt.Sprintf(`
+- sqlid:      %s
+- hostname:   %s
+- databases:  %s
+- tablename:  %s
+- sql_type:   %s
+- run_time:   %s
+- row_scan:   %s
+- sort_by:    %s
+- exec_plan:  %s
+- query_part: %s       
+- alert_time: %s 
+- slow_message: %s`,
   sqlid,
   Get_v(query,"host"),
   Get_v(query,"db"),
   Get_v(query,"table"),
   optp,
   Get_v(query,"ms")+"ms",
   Get_v(query,"doc_scan"),
   Get_v(query,"sort"),
   Get_v(query,"plan"),
   Get_v(query,"query_part"),
   now,line)
  
}



func Get_host(msg string) string{
     //Dec  3 19:40:51 webdb-3 mongod.5703[39525]: [conn446686]
     par:=`(?P<hostname>\S+)\s+(?P<pre>mongod\.\d+\[\d+\]\: \[conn\d+\])`
     rst:=Regexp_map(par,msg)
     if v,ok:=rst["hostname"];ok{
        return v
     }
     return ""
}


func Slow_f(msg string,ltm int) (bool,string){
     par:=`\s+(?P<ms>\d+)ms$`
     rst:=Regexp_map(par,msg) 
     if v,ok:=rst["ms"];ok{
        if qtm,err:=strconv.Atoi(v);err==nil && qtm >ltm {
           return true,v
           }  
     }
     return false,""
}



func md5V3(str string) string {
    w := md5.New()
    io.WriteString(w, str)
    md5str := fmt.Sprintf("%x", w.Sum(nil))
    return md5str
}


func get_query(q string) string{
     qx:=Qt{Allstring:q,Index:0,Stat:0,Lstat:0,Indic:0,Debug:false} 
     qx.Scan()
     r:=regexp.MustCompile(`"`) 
     return r.ReplaceAllString(string(qx.Qstring.Bytes()),"")
}

func gen_qid(query map[string]string) string {
     optp:=Get_v(query,"optp")
     if optp=="ddl"{return Get_v(query,"query_part")}
     host:=Get_v(query,"host")
     db:=Get_v(query,"db")
     tab:=Get_v(query,"table")
     query_sql:=Get_v(query,"query_part")
     var qid string
     if query_sql!=""{
        query_id:=get_query(query_sql)
        qid=fmt.Sprintf(`%s.%s.%s.%s.%s`,host,db,tab,optp,query_id )
     }else{
        qid=fmt.Sprintf(`%s.%s.%s.%s`,host,db,tab,optp)
     }
     qid=md5V3(qid)
     return qid 
}





