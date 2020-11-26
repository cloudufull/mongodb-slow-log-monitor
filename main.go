package main


import (
        "flag"
        "fmt"
        "os"
        "time"
        "io"
//        "log"
//        "io/ioutil"
        "bufio"
        "strconv"
        "strings"
        "regexp"
        "mslow/parse"
)



var message_log = flag.String("msg", "/var/log/message", "mongodb slow log file")
var tm_formart = flag.String("time_format", "2006-01-02T15:04:05@1", "slow log time format and include columns number")
var tm_start =flag.String("start","","time from ,format like '2006-01-02 15:04:05'")
var tm_end =flag.String("end","","time end ,format like '2006-01-02 15:04:05' ")
var slow_tm = flag.Int("st", 0, "slow query threshold ")
var flasttm =flag.Int64("last",0,"last min")
var fl =flag.Int("follow",0,"follow message log")
var bs =flag.Int("follow_batchsize",5000,"read follow message until $follow_batchsize number line or to the end")
var mail_from =flag.String("mail_from","","alert mail address  'from' part ")
var mail_to =flag.String("mail_to","","alert mail address  'to' part , many address use ',' divide")
var dc =flag.String("dc",""," alert from which dc")
var acount=flag.Int("acount",3,"auto ack alert")
var ack=flag.String("ack","","acknowledge qid ,not alert any more")
var trange=flag.String("tr","","ack time range")
var comment =flag.String("C","","ack comment ")
var rpt =flag.Bool("rpt",false,"print report range fstart to fend ")
var qid =flag.String("qid","","report by qid ") 
var sort =flag.String("sort","avg","repoty sort by avg or sort by cnt ")
var hst =flag.String("host","","report print by host ")
var debug =flag.Bool("v",false,"print more info ")
var rptd = flag.Int("rptd",-1,"report auto ack sql detail to mail within N days ")
var mongo_log = flag.Bool("pm",false," log producted by mongd not syslog  ")



func main() {
     // options
        flag.Usage = func() {
                fmt.Fprintf(os.Stderr, "use analyze mongodb slow log when  systemLog.destination:syslog\n")
                fmt.Fprintf(os.Stderr, "Usage: %s [options ...]\n", os.Args[0])
                fmt.Fprint(os.Stderr, "\n")
                fmt.Fprint(os.Stderr, "Options:\n")
                fmt.Fprint(os.Stderr, "\n")
                fmt.Fprint(os.Stderr,`         -msg   default /var/log/message `,"\n") 
                fmt.Fprint(os.Stderr,`         -time_format   default "2006-01-02T15:04:05@[1]" messge log time format
                        eg:  when message log  time info like this "Dec 25 03:38:02 hn15 mongod.5703[378028]" 
                             we should use -time_format "Jan  2 15:04:05@0" ,  "@0" means   time info start from column 0    
                             eg1 : time format  2006-01-02T15:04:05.000-0700  <=----------=>  msg like : 2020-11-12T04:30:01.615+0800 

                             `,"\n")

                fmt.Fprint(os.Stderr,`         -start   default ""  appoint time begin ,format like '2006-01-02 15:04:05'  `,"\n") 
                fmt.Fprint(os.Stderr,`         -end     default ""  appoint time end ,format like '2006-01-02 15:04:05'  `,"\n") 
                fmt.Fprint(os.Stderr,`         -last    default 0, analyze within last min message log `,"\n") 
                fmt.Fprint(os.Stderr,`         -st      default 0, assign slow query threshold `,"\n") 
                fmt.Fprint(os.Stderr,`         -follow  default 0, when 1  batch read message log from log head 
                             when 2 batch read message log from the log's end`,"\n") 
                fmt.Fprint(os.Stderr,"\n")
                fmt.Fprint(os.Stderr,`         -follow_batchsize  default 5000, when use -follow ,batch read number `,"\n") 
                fmt.Fprint(os.Stderr,"\n")
                fmt.Fprint(os.Stderr,"Report :\n")
                fmt.Fprint(os.Stderr,`         -rpt     default false, make report beteew start and end `,"\n") 
                fmt.Fprint(os.Stderr,`                  -sort    default "avg",  sort by avg or cnt `,"\n") 
                fmt.Fprint(os.Stderr,`                  -host    default "",  report print by host  also used to ack option`,"\n") 
                fmt.Fprint(os.Stderr,`                  -qid    default "",  report print by qid`,"\n") 
                fmt.Fprint(os.Stderr,`         -rptd    default 3,  report auto ack sql detail to mail within N day （need mail_from mail_to dc）`,"\n")
                fmt.Fprint(os.Stderr,"\n")
                fmt.Fprint(os.Stderr,"Alert Mail :\n")
                fmt.Fprint(os.Stderr,"\n")
                fmt.Fprint(os.Stderr,`         -mail_from  default "", if assign mail_from ,will send alert message about slow query log, also need -mail_to -dc -acount  `,"\n") 
                fmt.Fprint(os.Stderr,`         -mail_to    default "",  mail will be sended to  muti use "," Separate, `,"\n") 
                fmt.Fprint(os.Stderr,`         -dc    default "",  mail from which dc, `,"\n") 
                fmt.Fprint(os.Stderr,`         -acount    default 3,  report how many times alert before auto ack `,"\n") 
                fmt.Fprint(os.Stderr,"\n")
                fmt.Fprint(os.Stderr,"Alert ack :\n")
                fmt.Fprint(os.Stderr,"\n")
                fmt.Fprint(os.Stderr,`         -ack    default "",  acknowledge qid ,not alert any more 
                 eg :  -ack h -host host001 -tr 0-5@12-14@20-23   # acknowledge all qid during 0h - 5h and 12-14h and 20-23 
                       -ack xxlen32xx    # acknowledge  qid "xxlen32xx"
                       -ack xxlen32xx -C "normal sql"    # acknowledge  qid "xxxxxxxxxxxxlen32xx"
                       -ack list                   # show all acknowledged qid  `,"\n") 
                fmt.Fprint(os.Stderr,`                 -tr    default "",  acknowledge time range by host  `,"\n") 
                fmt.Fprint(os.Stderr,`                 -C    default "",  ack reason   `,"\n") 
                fmt.Fprint(os.Stderr,`         -v    default false,  print more info`,"\n") 
                fmt.Fprint(os.Stderr,`         -pm   default false,  not use syslog message`,"\n") 
        }
        flag.Parse()
        

         
        // time format 
        fm:=strings.Split(*tm_formart,"@")[0]
        col,err:=strconv.Atoi(strings.Split(*tm_formart,"@")[1])
        if err!=nil{
           fmt.Println(err,"<<---------------")
           panic("time format error")
        }

        // get flag
        var start_flag int64 =0
        var end_flag int64 =0
        var t time.Time
        if *tm_start!=""{
           //t, err = time.Parse("2006-01-02 15:04:05", *tm_start)
           //t, err = time.ParseInLocation(TIME_LAYOUT, *tm_start, localt)
           t,err = parse.Ftime(*tm_start)
           if err != nil {
              panic("start time format error") 
           }
           fmt.Println(*tm_start,t.Unix())
           start_flag=t.Unix()
        }

        if *tm_end!=""{
           //t, err = time.Parse("2006-01-02 15:04:05", *tm_end).In(localt)
           //t, err = time.ParseInLocation(TIME_LAYOUT, *tm_end, localt)
           t,err = parse.Ftime(*tm_end)
           if err != nil {
              panic("end time format error:") 
           }
           end_flag=t.Unix()
           if *tm_start==""{
              panic("can not only provide end_time need start_time:") 
           } 
        }
        if *flasttm!=0{
           localt, _ := time.LoadLocation("Local")
           start_flag=time.Now().In(localt).Unix() - 60**flasttm 
        }
        
        var mp parse.Mail_info
        sqlite:=parse.Get_db(*mail_from,*mail_to,*dc,*debug)
        if mp.Mail_from!="" && mp.Mail_to!="" && mp.DC!=""{
           sqlite.Mail=&mp
        }
        if *ack!=""{
            if *ack=="list"{
               sqlite.Find_all_ack() 
            }else{
               sqlite.Save_ack(*ack,int(time.Now().Unix()),*hst,*trange,*comment)
            }
            return 
        }  
        sqlite.Check_table()
        if *rpt {
           fmt.Println(start_flag,"---",end_flag)
           sqlite.Report(int(start_flag),int(end_flag),0,*hst,*sort,*qid) 
           return
        }
        if *rptd >= 0 {
           if *debug{fmt.Println(*rptd)} 
           sqlite.Report_daily(*mail_from,*mail_to,*dc,*rptd)
           return
       }

        var lastpos int64
        file, err := os.OpenFile(*message_log, os.O_RDWR, 0666)
        if *fl==1{
            lastpos=sqlite.Get_fpos(*message_log)
            _,err=file.Seek(0,os.SEEK_END)
            cur_pos,err:=file.Seek(0, os.SEEK_CUR)
            parse.Checkerr(err)
            if int(cur_pos)<int(lastpos){
               lastpos=0
            }
            _,err=file.Seek(lastpos,0)
            if *debug{fmt.Println(lastpos)}
        }

        if *fl==2{
            lastpos=sqlite.Get_fpos(*message_log)
            _,err=file.Seek(0,os.SEEK_END)
        }
        

        parse.Checkerr(err)
        buf := bufio.NewReader(file)
      
        readline:=0
        for {
            line, err := buf.ReadString('\n')
            if *fl>0{
                readline=readline+1
                if readline>=*bs{
                   pos,err:=file.Seek(0, os.SEEK_CUR)  
                   parse.Checkerr(err)
                   pos=pos-int64(buf.Buffered())
                   sqlite.Update_fpos(*message_log,pos)
                   break
                }
            }
            if err != nil {
                if err == io.EOF {
                   pos,err:=file.Seek(0, os.SEEK_CUR)  
                   parse.Checkerr(err)
                   sqlite.Update_fpos(*message_log,pos)
                   if *debug{fmt.Println(pos)}
                   break
                } else {
                    fmt.Println("Read file error!", err)
                    return
                }
            }
            if n,_:=regexp.MatchString(` mongod\.\d+\[`,line);!n{
                  if *mongo_log{
                     re:=regexp.MustCompile(` \[conn\d+\] `)
                     sidx:=re.FindStringIndex(line) 
                     if sidx!=nil{
                      line=fmt.Sprintf("%smongod.5703[1111111]:%s",line[0:sidx[0]],line[sidx[0]:])
                     }
                  }else{
                      continue
                  }
            }
            line = strings.TrimSpace(line)
            z:=strings.Split(strings.Trim(line,"\n")," ")   
            if len(z)==0||len(line)==0{
               continue 
            }
            var charge_fm string 
            tmstr:=strings.Join(z[col:]," ")[:len(fm)]
            charge_fm=fm
            if n,_:=regexp.MatchString("2006",fm);!n{
               tmstr=strings.Join(z[col:]," ")[:len(fm)]+" "+fmt.Sprintf(time.Now().Format("2006"))
               fmt.Println(n);
               charge_fm=charge_fm+" 2006"
            }
            //fmt.Println(charge_fm," <=----------=> ",tmstr);
            t, _  := time.ParseInLocation(charge_fm, tmstr,time.Local)
            //fmt.Println(t,t.Unix(),end_flag,charge_fm," <=> ",tmstr)
            //localt, _ := time.LoadLocation("Local")
            //fmt.Println(t.Format("2006-01-02 15:04:05.000+0700")," <=>",tmstr,"<=>",charge_fm)
            if end_flag!=0 && t.Unix()>end_flag{ break }
            if start_flag!=0{
               if t.Unix()>start_flag{
                 sqlite.Output(line,*slow_tm,*acount,t)
                }
            }else{
                sqlite.Output(line,*slow_tm,*acount,t)
            }
        }
        //fmt.Println(sqlite.Mail.Mail_from,"<-------")

        for _,v :=range sqlite.Mail.Mailbox { 
            msg:=strings.Join(v,"\n")
            if *debug{fmt.Println(msg)}
            if sqlite.Mail.Mail_from!="" && sqlite.Mail.Mail_to!="" && sqlite.Mail.DC!=""{
               parse.Mail(sqlite.Mail.Mail_from,sqlite.Mail.Mail_to,sqlite.Mail.DC,msg)
            }
        }
         
        defer func(){
           if err:=recover();err!=nil{ 
              fmt.Println(" cache exception (panic)..!! ",err)
              pos,err:=file.Seek(0, os.SEEK_CUR)
              parse.Checkerr(err)
              sqlite.Update_fpos(*message_log,pos)
              fmt.Println(pos)
           }
        }() 
         

        
}
