# mongodb-slow-log-monitor


# a tool for Monitor mongodb slow query through rsyslog
  
  before use this you should config you mongodb like this:

```
    systemLog:
    destination: syslog
    verbosity: 0
    syslogFacility: "user"
```   

### use

```
export GO111MODULE=on 
export GOPROXY=https://goproxy.io
go build -o go_mslow main.go 

#]./go_mslow --help

use analyze mongodb slow log when  systemLog.destination:syslog


Options:

         -msg   default /var/log/message 
         -time_format   default "2006-01-02T15:04:05@[1]" messge log time format
                        eg:  when message log  time info like this "Dec 25 03:38:02 hn15 mongod.5703[378028]" 
                             we should use -time_format "Jan  2 15:04:05@0" ,  "@0" means   time info start from column 0    
         -start   default ""  appoint time begin ,format like '2006-01-02 15:04:05'  
         -end     default ""  appoint time end ,format like '2006-01-02 15:04:05'  
         -last    default 0, analyze within last min message log 
         -st      default 0, assign slow query threshold 
         -follow  default 0, when 1  batch read message log from log head 
                             when 2 batch read message log from the log's end

         -follow_batchsize  default 5000, when use -follow ,batch read number 

Report :
         -rpt     default false, make report beteew start and end 
                  -sort    default "avg",  sort by avg or cnt 
                  -host    default "",  report print by host  also used to ack option
                  -qid    default "",  report print by qid
         -rptd    default 3,  report auto ack sql detail to mail within N day ��need mail_from mail_to dc��

Alert Mail :

         -mail_from  default "", if assign mail_from ,will send alert message about slow query log, also need -mail_to -dc -acount  
         -mail_to    default "",  mail will be sended to  muti use "," Separate, 
         -dc    default "",  mail from which dc, 
         -acount    default 3,  report how many times alert before auto ack 

Alert ack :

         -ack    default "",  acknowledge qid ,not alert any more 
                 eg :  -ack h -host host001 -tr 0-5@12-14@20-23   # acknowledge all qid during 0h - 5h and 12-14h and 20-23 
                       -ack xxlen32xx    # acknowledge  qid "xxlen32xx"
                       -ack xxlen32xx -C "normal sql"    # acknowledge  qid "xxxxxxxxxxxxlen32xx"
                       -ack list                   # show all acknowledged qid  
                 -tr    default "",  acknowledge time range by host  
                 -C    default "",  ack reason   
         -v    default false,  print more info

    
```

### example

1. ./go_mslow -msg /var/log/messages -time_format "Jan  2 15:04:05@0" -st 1000 -follow  1 -follow_batchsize 5000 -mail_from "mg@alert.com" -mail_to "xxxx@pd.com,364263756@qq.com" -dc "beijing" -acount 3
       
alert message like this:

```	 
	 +- sqlid:      97eb220da18c9b3676ef98859936931e
         +- hostname:   xxxhos
         +- databases:  dbN
         +- tablename:  tabe1
         +- sql_type:   find
         +- run_time:   1561ms
         +- row_scan:   1
         +- sort_by:    
         +- exec_plan:  COLLSCAN
         +- query_part: { name: "sgxxxxid" }       
         +- alert_time: 2020-04-07 09:45:41 
         +- slow_message: Apr  7 09:45:41 webdb-5 mongod.5702[67935]: [conn913382] command dbN.tabe1 command: find { find: "tabe1", filter: { name: "sgxxxxid" }, 
	                  projection: { _id: 0, __v: 0 }, limit: 1, singleBatch: true, batchSize: 1 } planSummary: COLLSCAN keysExamined:0 docsExamined:1 cursorExhausted:1 
			  keyUpdates:0 writeConflicts:0 numYields:1 nreturned:1 reslen:190 locks:{ Global: { acquireCount: { r: 4 } }, Database: { acquireCount: { r: 2 } }, 
			  Collection: { acquireCount: { r: 2 } } } protocol:op_query 1561ms
```
     
2.   ./go_mslow -rpt -start '2020-04-09 00:00:00' -host host15


![image](https://github.com/cloudufull/mongodb-slow-log-monitor/blob/master/11.png)
          
	  
     ./go_mslow  -rptd 3 -mail_from "mg@alert.com" -mail_to "xxxx@pd.com,364263756@qq.com" -dc "beijing" 
	   
![image](https://github.com/cloudufull/mongodb-slow-log-monitor/blob/master/22.png)

	   