package parse 

import (
       "database/sql"
       "fmt"
       _ "github.com/mattn/go-sqlite3"
       "time"
       "regexp"
       "strings"
       "strconv"
       "text/template"
       "bytes"
       "os"
       "gonum.org/v1/plot"
       "gonum.org/v1/plot/plotter"
       "gonum.org/v1/plot/plotutil"
       "gonum.org/v1/plot/vg" 
       "github.com/vanng822/go-premailer/premailer" 
)





func (sqlite *DB_save)Output(line string,slow_tm int,account int,ltm time.Time,hst string) {
     //fmt.Println("===================================================")
     report:=true 
     mp:=Change_type(line)
     //hst:=Get_host(line)
     var msg string
     var qid string
     if _,tm,host,trange:=sqlite.Find_ack("#host#"+hst);tm>0{
        if host!="" && host==hst{
           if sqlite.Debug{fmt.Println("-----------host-------skip:   ",host)}
           nowhour:=ltm.Hour()
           if trange==""{
              report=false 
           }else{
              aqlst:=strings.Split(trange,"@")
              for _,i:=range aqlst {
                  s:=strings.Split(i,"-")[0]
                  e:=strings.Split(i,"-")[1]
                  st,_:=strconv.Atoi(s)
                  et,_:=strconv.Atoi(e)
                  if nowhour>st && nowhour<et{
                     report=false 
                  }
              }
           }
           defer func(){
            if err:=recover();err!=nil{
                fmt.Println(" cache exception (panic)..!! ",err)
                }
           }()

        }
       
     }

     if report {      

          if optp:=Get_v(mp,"optp");optp=="build Index" || optp=="exception"{
             t:=ltm.Format("2006-01-02 15:04:05") 
             if build_index,_:=regexp.MatchString("build index",line);build_index{
                rst:=Regexp_map(`(?P<loc> \w+\.\w+)(?P<key> properties):(?P<value>.+)`,line) 
                ddl_content:=Get_v(rst,"loc")+Get_v(rst,"key")+Get_v(rst,"value")
                qid=md5V3(ddl_content)
                if _,tm,_,_:=sqlite.Find_ack(qid);tm>0{
                   if sqlite.Debug{fmt.Println("**skip  ack build Index  sqlid ",qid," : ",ddl_content)}
                   return      
                }
             }else{
                if len(Get_v(mp,"query_part"))>0{
                   qidstr:=Get_v(mp,"table")+Get_v(mp,"db")+hst+Get_v(mp,"query_part")
                   qid=md5V3(qidstr)
                }else{
                   qid=md5V3(Get_v(mp,"query_part"))
                   }
             }
             if qid !="" {
                emp:=&Save_alert_st{
                qtm:int(ltm.Unix()),
                atm:int(time.Now().Unix()),
                etime:0,
                qid:qid,
                db:Get_v(mp,"db"),
                table:Get_v(mp,"table"),
                host:hst,
                optp:optp,
                qmessage:Regexp_map(`(?P<pre>\s\[conn\d+\]\s)(?P<msg>.+)`,line)["msg"]}
                sqlite.Save_alert(emp)
                // // //   --   auto ack alert  -- // // //
                if sqlite.Find_alert_count(qid,account){
                   if _,tm,_,_:=sqlite.Find_ack(qid);tm==0 {
                      amsg:=fmt.Sprintf("auto ack after %d times",account)
                      if sqlite.Debug{fmt.Println("**sqlid",qid," auto ack after ",account," times ")}
                      sqlite.Save_ack(qid,int(time.Now().Unix()),"","",amsg)
                   }
                }
                // find ack
                if _,tm,_,_:=sqlite.Find_ack(qid);tm>0 {
                   if sqlite.Debug{fmt.Println("**skip ack sqlid ",qid," : ",emp.qmessage)}
                   report=false
                }
             }
             msg=fmt.Sprintf(`
+- hostname:   %s
+- sql_type:   %s
+- alert_time: %s
+- ddlqid:%s
+- slow_message: %s`, hst,optp,t,qid,line)
                
            }else if ok,ms:=Slow_f(line,slow_tm);ok&&len(mp)>0{
               mp["ms"]=ms
               mp["host"]=hst
               qid,msg=sqlite.Format_output(mp,line,account,ltm)
          }
          
          if msg!="" && report { 
             if v,ok:=sqlite.Mail.Mailbox[qid];ok{
                v=append(v,"               : "+line)
                sqlite.Mail.Mailbox[qid]=v
             }else{
                 lst:=make([]string,1)
                 lst=append(lst,msg)
                 sqlite.Mail.Mailbox[qid]=lst
             }
         }
    }
}



func Get_db(mf string,mt string ,dc string,v bool) *DB_save {
     dbs,err := sql.Open("sqlite3", "./foo.db")
     Checkerr(err)
     mlbox:=make(map[string][]string) 
     db:=&DB_save{db:dbs,Debug:v}
     db.Mail=&Mail_info{
           Mail_from:mf,
           Mail_to:mt,
           DC:dc,
           Mailbox:mlbox}
       return db 

}



func Gettimerange(tm int64) (int64,int64) {
    now := time.Unix(tm,0) 
    stimestamp := now.Unix() - int64(now.Second()) - int64((60 * now.Minute()))
    mtimestamp:=stimestamp+60*30
    etimestamp:=stimestamp+60*60
    if tm<mtimestamp{
       return stimestamp,mtimestamp
    }else{
       return mtimestamp,etimestamp
       }
}


func Gettime_report(last_day int) (int64,int64) {
     now := time.Now()
     n_days_ago := now.Unix() - int64(now.Second()) - int64((60 * now.Minute()))-int64(last_day)*int64(60*60*24)
     return n_days_ago,now.Unix()

}




func (sqlite *DB_save)Check_table(){
     var init_sql map[string]string=make(map[string]string)
     init_sql["qidtime"]=`
         create table qidtime(qid VARCHAR(32),long int,qtm int,atm int);
         create index idx_qqt on qidtime(qid,qtm);
         create index idx_qat on qidtime(qid,atm);
          `
     
     init_sql["qidsql"]=`
        create table qidsql(qid VARCHAR(32),sql TEXT,db VARCHAR(32),tab VARCHAR(32),host VARCHAR(16),optp VARCHAR(8)); 
        create unique index idx_qid on qidsql(qid);
     `
     
     init_sql["qidack"]=`
        create table qidack(qid VARCHAR(32),acktime int,host VARCHAR(64),trange VARCHAR(64),comment VARCHAR(128) default ""); 
        create unique index idx_qack on qidack(qid);
        create unique index idx_hack on qidack(host);
     `

     init_sql["logpos"]=`
        create table logpos(fl VARCHAR(200),pos int64);
        create index idxfl on logpos(fl);
     `

     init_sql["conn_dic"]=`
        create table conn_dic(host varchar(200),connid VARCHAR(200),client_addr VARCHAR(200),tm int);
        create index idx_dic_con on conn_dic(connid,tm);
        create index idx_dic_tm  on conn_dic(tm);
        create index idx_dic_tm  on conn_dic(host,connid,tm);
     `
      
     tablst:=[5]string{"qidtime","qidsql","qidack","logpos","conn_dic"}
     var cnt int64
     for _,tab:=range tablst{ 
         err := sqlite.db.QueryRow(`SELECT COUNT(*) as cnt FROM sqlite_master WHERE type='table' AND name = ?`,tab).Scan(&cnt)
         Checkerr(err)
         if cnt<=0{
               _,err=sqlite.db.Exec(init_sql[tab])
               if err == sql.ErrNoRows{ 
                  Checkerr(err)
               }
               //fmt.Println(init_sql[tab])
            } 
        
    }
}


func (sqlite *DB_save)Save_conn_dict(host string,cid string,cip string,tm int) {
      _, err := sqlite.db.Exec(`insert into conn_dic(host,connid,client_addr,tm) values (?,?,?,?)`,host,cid,cip,tm)
      if err != sql.ErrNoRows{
         Checkerr(err)
      }
}


func (sqlite *DB_save)Pop_cid(cid string,host string){
      _, err := sqlite.db.Exec(`delete from conn_dic where connid=? and host=?`,cid,host) 
      Checkerr(err)
}



func (sqlite *DB_save)Get_client_ip(cid string,hst string) string{
      var cip string
      err := sqlite.db.QueryRow(`select client_addr from conn_dic where connid=? and host=? order by tm desc limit 1 `,cid,hst).Scan(&cip) 
      switch {
         case err == sql.ErrNoRows:
             if sqlite.Debug{fmt.Println("No client ip found for cid:",cid)}
         case err != nil:
             fmt.Println(`Get_client_ip`,err)
         default:
            return cip 
      } 
      return ""
}



func (sqlite *DB_save)Save_alert(qmap *Save_alert_st) {
     _, err := sqlite.db.Exec("insert into qidtime(qid ,long ,qtm ,atm ) values(?,?,?,?)",qmap.qid,qmap.etime,qmap.qtm,qmap.atm)
     Checkerr(err)
     var count int64
     _ = sqlite.db.QueryRow(`select count(*)  FROM qidsql where qid=?`, qmap.qid).Scan(&count)
     if count<1{
        _, err= sqlite.db.Exec("insert into qidsql(qid ,sql,db,tab,host,optp) values(?,?,?,?,?,?)",qmap.qid,qmap.qmessage,qmap.db,qmap.table,qmap.host,qmap.optp)
        Checkerr(err)
     }
     
}

func (sqlite *DB_save)Save_ack(qid string,qtm int,h string ,tr string,comment string)  {
     var count int64
     if qid=="h" && h!=""{
        qid=RandString(32)
        fmt.Println("save null qid switch ",qid,"for host",h)
     }
     if h=="" {    
       _ = sqlite.db.QueryRow(`select count(*)  FROM qidack where qid=?`, qid).Scan(&count)
     }else{
       _ = sqlite.db.QueryRow(`select count(*)  FROM qidack where qid=? and host=?`, qid,h).Scan(&count)
     }
     var err error
     if count<1{
        if h==""{
           _, err= sqlite.db.Exec("insert into qidack(qid,acktime,comment) values(?,?,?)",qid,qtm,comment)
        }else{
           _, err= sqlite.db.Exec("insert into qidack(qid,acktime,host,trange,comment) values(?,?,?,?,?)",qid,qtm,h,tr,comment)
        }
        Checkerr(err)
     }
}

func (sqlite *DB_save)Find_alert_count(qid string,tms int) bool {
     var count int64 
     err := sqlite.db.QueryRow("select count(*) as cnt from qidtime where qid=? ", qid).Scan(&count)
     Checkerr(err)
     if int(count)>tms{
        return true 
      }else{
        return false 
      }
}


func (sqlite *DB_save)Find_ack(id string) (string, int64, string, string) {
     var qid string 
     var acktime int64
     var host string
     var trange string
     var err error
     if strings.HasPrefix(id,"#host#"){
        id=strings.Replace(id,"#host#","",-1)
        err = sqlite.db.QueryRow("select qid,acktime,host,trange from qidack where host=? ", id).Scan(&qid,&acktime,&host,&trange)
        if err!= sql.ErrNoRows{Checkerr(err)}
     }else{
        err = sqlite.db.QueryRow(`select qid,acktime from qidack where qid=? `, id).Scan(&qid,&acktime)
     }
     if err==nil && err!= sql.ErrNoRows{
        return qid,acktime,host,trange

     }
     return qid,acktime,host,trange
     
}


func (sqlite *DB_save)Find_all_ack(){
     rows,err:= sqlite.db.Query(`select qid,acktime,ifnull(host,"."),ifnull(trange,"."),ifnull(comment,".")  from qidack `)
     defer rows.Close()
     fmt.Printf("%32s|%16s\n","sqlid","acktime")
     fmt.Println("-------------------------------------------")
     for rows.Next() {
         var qid string
         var ack int64 
         var host string
         var trange string
         var  comment string
         err = rows.Scan(&qid, &ack,&host,&trange,&comment)
         Checkerr(err) 
         fmt.Printf("%32s|%16s|%5s|%5s|%5s\n",qid,time.Unix(ack,0).Format("2006-01-02 15:04:05"),host,trange,comment)
      }
}



func (sqlite *DB_save)Find_detail(id string) map[string]string {
     var qid string
     var sql string
     var db string
     var table string
     var host string 
     var optp string
     err := sqlite.db.QueryRow("select * from qidsql where qid=?", id).Scan(&qid,&sql,&db,&table,&host,&optp)
     rmp:=make(map[string]string)
     if err==nil{ 
        rmp["qid"]=qid
        rmp["sql"]=sql
        rmp["db"]=db 
        rmp["table"]=table
        rmp["host"]=host 
        rmp["optp"]=optp
        return rmp
     }
     Checkerr(err)
     return rmp
}

// report sql query

func (sqlite *DB_save)Report(stm int,etm int,wide int,host string,sortby string,cqid string) {
     var mintm int64
     var maxtm int64
     var sql string
     if etm==0 { etm=9999999999}
     sortc:="avg"
     if sortby!="avg"{sortc="cnt"}
     
     //fmt.Printf("select min(qtm),max(qtm) from qidtime where qtm >%d and qtm<%d \n", stm,etm)
     err := sqlite.db.QueryRow("select min(qtm),max(qtm) from qidtime where qtm >=? and qtm<=?", stm,etm).Scan(&mintm,&maxtm)
     if cqid==""{
        sql=fmt.Sprintf(`select qid,avg(long) as avg ,count(*) as cnt from qidtime where qtm >=%d and qtm<=%d group by qid order by %s desc`,stm,etm,sortc)
     }else{
        sql=fmt.Sprintf(`select qid,avg(long) as avg ,count(*) as cnt,max(long) as max,min(long) as min from qidtime where qtm >=%d and qtm<=%d and qid="%s" group by qid order by %s desc`,stm,etm,cqid,sortc)
      }


     rows,err:= sqlite.db.Query(sql)

     defer rows.Close()
     // get wide size
     if wide==0{wide=Get_terminle_size()}
     fmt.Println()
     TIME_LAYOUT:="2006-01-02 15:04:05"
     localt, _ := time.LoadLocation("Asia/Shanghai")

     fmt.Printf(" ||||| from [%s] to [%s] < %d --> %d > ||||| \n",time.Unix(mintm,0).In(localt).Format(TIME_LAYOUT),time.Unix(maxtm,0).In(localt).Format(TIME_LAYOUT),mintm,maxtm)

     fmt.Println("-----------------------------------------------------------------------------------------------------------------------------")
     if cqid==""{
        fmt.Printf("%32s|%10s|%19s|%21s|%15s|%10s|%10s\n","sqlid","host","database","table","type","avg time(ms)","count")
     }else{
        fmt.Printf("%10s|%19s|%21s|%15s|%10s|%10s|%10s|%10s\n","host","database","table","type","avg time(ms)","count","max","min")
     }

     fmt.Println("-----------------------------------------------------------------------------------------------------------------------------")
     qidlst:=make(map[string]string)
     for rows.Next() {
         var qid string
         var avg float64
         var cnt int64
         var max int64
         var min int64
         if cqid ==""{
            err = rows.Scan(&qid, &avg,&cnt)
         }else{
            err = rows.Scan(&qid, &avg,&cnt,&max,&min)
         } 
         Checkerr(err)
         
         if qid!="" {
            dmp:=sqlite.Find_detail(qid)
            if cqid!=""{  
               msg:=fmt.Sprintf("%10s|%19s|%21s|%15s|%10f|%10d|%10d|%10d\n",Get_v(dmp,"host"),Get_v(dmp,"db"),Get_v(dmp,"table"),Get_v(dmp,"optp"),avg,cnt,max,min) 
               if host!="" {
                 if host==Get_v(dmp,"host"){
                    fmt.Printf(msg)
                  }
               }else{
                  fmt.Printf(msg)
               }
            }else{
                msg:=fmt.Sprintf("%32s|%10s|%19s|%21s|%15s|%10f|%10d\n",qid,Get_v(dmp,"host"),Get_v(dmp,"db"),Get_v(dmp,"table"),Get_v(dmp,"optp"),avg,cnt)
                if host!="" {
                   if host==Get_v(dmp,"host"){
                      fmt.Printf(msg)
                      qidlst[qid]=Get_v(dmp,"sql")
                   }
                }else{
                   fmt.Printf(msg)
            //       qidlst[qid]=Get_v(dmp,"sql")
                }
            }
            if host==""{qidlst[qid]=Get_v(dmp,"sql")}
         }
         
     }
     fmt.Printf("\n======================\n")
     fmt.Printf("|| slow query list ||:\n")
     fmt.Printf("----------------------\n\n")
     fmt.Println("             sqlid               |    sqlmessage       ")
     fmt.Println("-----------------------------------------------------------------")
     vlen:=wide-36
     for k,v:=range qidlst{
         fmt.Printf("%1s : %s\n",k,Split_msg(v,vlen))
     }
     if cqid!=""{
        fmt.Println("\n===================\n")
        fmt.Printf("%9s|%17s|%17s\n","query time ","occurd time ","alert time")
        sql=fmt.Sprintf(`select long ,qtm ,atm from qidtime where qtm >=%d and qtm<=%d and qid="%s" order by atm asc`,stm,etm,cqid)
        rows,err:= sqlite.db.Query(sql)
        Checkerr(err)
        for rows.Next() {
            var qlong int64
            var qtm int64
            var atm int64
            err = rows.Scan(&qlong, &qtm,&atm)
            fmt.Printf("%11d|%16s|%16s\n",qlong,time.Unix(qtm,0).In(localt).Format(TIME_LAYOUT),time.Unix(atm,0).In(localt).Format(TIME_LAYOUT))
        }

     }
     Checkerr(err)
}


func (sqlite *DB_save)Report_daily(from string,to string,dc string,last_day int){
     folderPath:="./image"
     if _, err := os.Stat(folderPath); os.IsNotExist(err) {
        os.Mkdir(folderPath, 0755) 
        os.Chmod(folderPath, 0755)
	}
     start,end:=Gettime_report(last_day)
     if sqlite.Debug{fmt.Println("report from ",start," to ",end)}   
     sql:=fmt.Sprintf(`select qid,acktime from qidack where acktime>=%d`,start)
     rows,err:= sqlite.db.Query(sql)
     Checkerr(err)
     clst:=make([]map[string]string,0,10)
     image_map:=make(map[string]string)
     for rows.Next() {
         qidmap:=make(map[string]string)
         var qid string 
         var atm int64
         err = rows.Scan(&qid, &atm)
         var cnt int64
         sql=fmt.Sprintf(`select count(*) as cnt from qidtime where qtm >=%d and qtm<=%d and qid="%s"`,start,end,qid) 
         err =sqlite.db.QueryRow(sql).Scan(&cnt)
         if cnt==0{ continue }
         dmp:=sqlite.Find_detail(qid)
         qidmap["qid"]=qid
         qidmap["sql"]=Get_v(dmp,"sql")
         qidmap["host"]=Get_v(dmp,"host")
         qidmap["db"]=Get_v(dmp,"db")
         qidmap["tab"]=Get_v(dmp,"table")
         qidmap["tp"]=Get_v(dmp,"optp")
         //query avg max min count
         var min int64
         var max int64
         var avg float64
           
         sql=fmt.Sprintf(`select avg(long) as avg ,count(*) as cnt,max(long) as max,min(long) as min from qidtime where qtm >=%d and qtm<=%d and qid="%s"`,start,end,qid)
         err =sqlite.db.QueryRow(sql).Scan(&avg,&cnt,&max,&min)
             
         if err==nil{
            sql=fmt.Sprintf(`select long,qtm from qidtime where  qid="%s" order by qtm asc `,qid)
            row,e:=sqlite.db.Query(sql)
            Checkerr(e)
            xdt:=make([]float64,0,10)
            ydt:=make([]float64,0,10)
            
            for row.Next(){
                var lng int64
                var qtm int64 
                e=row.Scan(&lng,&qtm)
                t:=time.Unix(qtm,0)
                xdt=append(xdt,float64(t.Unix())) 
                ydt=append(ydt,float64(lng/1000)) 
            }
             
            tupan_path:="null"
            if len(xdt)>0{
               tupan_path=fmt.Sprintf("./image/%s.png",qid)
               dt:=&Dataset{Len:len(xdt),
                 TT:"query time list",
                 X:"time",
                 Y:"long(s)",
                 Xdt:xdt,
                 Ydt:ydt,
                 Imagepath:tupan_path}
               draw(dt)  
               qidmap["tupian"]=fmt.Sprintf(`<img src="cid:%s.png" width="100%" height="100%">`,qid)
               if tupan_path!="null"{
                  image_map[qid]=tupan_path
               }
            }
               
            Checkerr(err)
            qidmap["load"]=fmt.Sprintf("%f -- %d -- %d -- %d",avg,max,min,cnt)
            //fmt.Println(cnt,min,max,avg,qid,start,end)
            qidmap["ack"]=strconv.Itoa(int(atm))
            clst=append(clst,qidmap)
         }else{fmt.Println(err)}
       }

     if len(clst)>0{
         tt:=fmt.Sprintf("auto ack slow query review within %d days @ %s",last_day,dc)
         mm:=Mail_report{
         Title:tt,
         Sendlst:"",
         Tcontent:&clst}
            
         t := template.Must(template.New("anyname").Parse(Mail_Template))
         buf := new(bytes.Buffer) 
         err= t.Execute(buf, &mm)
         prem, err := premailer.NewPremailerFromBytes(buf.Bytes(), premailer.NewOptions())
         Checkerr(err)
         html, err := prem.Transform()
         if err != nil {
             fmt.Println("Executing template:", err)
         }
         mbox:=&Mbox{
               From:from,
               To:to,
               Image:image_map,
               Title:tt, 
               Msg:html}
         if sqlite.Debug{fmt.Printf(" from %s to %s \n    %#v \n",from,to,image_map)}   
         Mailimage(mbox)

    }
}




func draw(dp *Dataset) {
     t:=MyTicks{}
     if len(dp.Xdt)>16 {
        SetTicks(int64(16))
     }else{
        SetTicks(int64(len(dp.Xdt)))
     }

     localt, _ := time.LoadLocation("Asia/Shanghai")
     xticks := plot.TimeTicks{Format: "01.02\n15:04",Ticker:t,Time:plot.UnixTimeIn(localt)}
     p, err := plot.New()
     p.X.Tick.Marker = xticks 
     Checkerr(err) 
     p.Title.Text = dp.TT 
     p.X.Label.Text = dp.X 
     p.Y.Label.Text = dp.Y 
     pts := make(plotter.XYs, dp.Len)
     for i := range pts {
         pts[i].X = dp.Xdt[i] 
         pts[i].Y = dp.Ydt[i] 
     }
     err = plotutil.AddLinePoints(p,pts)
     Checkerr(err)
     if err==nil{
        err = p.Save(12*vg.Inch, 4*vg.Inch, dp.Imagepath)
        if err!=nil{fmt.Println(err)}
        //Checkerr(err)
     }
}



// follow message log pos save & get

func (sqlite *DB_save)Get_fpos(fl string) int64 {
     var pos int64
     err := sqlite.db.QueryRow("select pos from logpos where fl=?", fl).Scan(&pos)
     switch {
         case err == sql.ErrNoRows:
             fmt.Println("No pos found for log file ",fl)
         case err != nil:
             fmt.Println(`Get_fpos`,err)
         default:
            return pos
     }
     return 0
}

func (sqlite *DB_save)Update_fpos(fl string,pos int64) {
     p:=sqlite.Get_fpos(fl)
     if p==0{
        _,err:= sqlite.db.Exec("insert into logpos(fl,pos) values(?,?)",fl,pos)
        Checkerr(err) 
     }else{
        _,err:= sqlite.db.Exec("update logpos set pos=? where fl=?",pos,fl)
        Checkerr(err)
     }
}
      

//func main() {
//     sqlite:=get_db()
// ////test check_table
//     sqlite.check_table()
// ////Save_alert
//     mp:=&Save_alert_st{
//     qtm:1578551127, 
//     etime:2817,
//     qid:"597e791e7ac22beab335ccebe6f6844c",
//     db:"notificationServer",
//     table:"userdeviceintercepts",
//     host:"db3",
//     optp:"count",
//     qmessage:`command notificationServer.userdeviceintercepts command: count { count: "userdeviceintercepts", query: { game_id: "27" } } planSummary: COLLSCAN keyUpdates:0 writeConflicts:0 numYields:54443 reslen:62 locks:{ Global: { acquireCount: { r: 108888 } }, Database: { acquireCount: { r: 54444 } }, Collection: { acquireCount: { r: 54444 } } } protocol:op_query 2817ms`}
//     sqlite.Save_alert(mp)
// ////Save_ack
//     sqlite.Save_ack("123asdasda",1578551127)
// ////Find_alert_count
//     fmt.Println(sqlite.Find_alert_count("597e791e7ac22beab335ccebe6f6844c",1578551127,10))
// ////Find_ack
//     fmt.Println(sqlite.Find_ack("597e791e7ac22beab335ccebe6f6844c"))
// ////Find_detail
//     fmt.Println(sqlite.Find_detail("597e791e7ac22beab335ccebe6f6844c"))
//      

//}
