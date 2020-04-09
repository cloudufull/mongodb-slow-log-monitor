package parse

import (
        "fmt"
        "strconv"
        "runtime"
        "os"
        "os/exec"
        "regexp"
        "strings"
        "github.com/go-gomail/gomail"
        "time"
        "math/rand"
)





// RandString 生成随机字符串
func RandString(len int) string {
    r := rand.New(rand.NewSource(time.Now().Unix()))
    bytes := make([]byte, len)
    for i := 0; i < len; i++ {
        b := r.Intn(26) + 65
        bytes[i] = byte(b)
    }
    return string(bytes)
}


func Ftime(tm string) (time.Time, error) {
     TIME_LAYOUT:="2006-01-02 15:04:05"
     localt, _ := time.LoadLocation("Asia/Shanghai")
     return time.ParseInLocation(TIME_LAYOUT, tm, localt)    


}


func Checkerr(n error) {
     if n!=nil{
        _,file,line,_ := runtime.Caller(1)
        loc:=file+" : "+strconv.Itoa(line)
        fmt.Fprint(os.Stderr, "error Found !!: @ ",n,loc,"\n")
      }
}

func CheckWarning(n error) {
     if n!=nil{
        _,file,line,_ := runtime.Caller(1)
        loc:=file+" : "+strconv.Itoa(line)
        fmt.Fprint(os.Stderr, "Warning !!: @ ",n,loc,"\n")
      }
}

func Get_v(mp map[string]string , k string) string {
     if v,ok:=mp[k];ok{
        if _,x:=interface{}(v).(int);x{
           //return strconv.Itoa(v)
            return ""
        }
        return v
     }
     return ""
}


func Regexp_map(par string,msg string) map[string]string{
     result := make(map[string]string)
     re,_:= regexp.Compile(par)
     match := re.FindStringSubmatch(msg)
     if len(match)==0{return result}
     groupNames := re.SubexpNames()

     for i, name := range groupNames {
         if i != 0 && name != "" { // 第一个分组为空（也就是整个匹配）
             result[name] = match[i]
         }
     }
     return result
}

func Get_terminle_size() int {
     cmd := exec.Command("stty", "size")
     cmd.Stdin = os.Stdin
     out, err := cmd.Output()
     if err==nil{
        col:=strings.Trim(strings.Split(string(out)," ")[1],"\n")
        n,_:=strconv.Atoi(col)
        return n
     }
     return 0
}


func Split_msg(msg string,l int) string{
     sublst:=make([]string,3,10)
     head:="                                "
     c:=0
     for _,v:=range msg{
         if c==l{
            c=0
            sublst=append(sublst,"\n")
            sublst=append(sublst,head)
         }
         sublst=append(sublst,string(v))
         c=c+1
     }
     return strings.Join(sublst,"")
}


// mail 
func Mail(from string,to string ,dc string,msg string){
     tos:=strings.Split(to,",")
     hd:=dc+"mongodb slow log review "
     m := gomail.NewMessage()
     m.SetHeader("From", from)
     m.SetHeader("To", tos...)
     m.SetHeader("Subject", hd)
     m.SetBody("text/plain", msg)

     d := gomail.Dialer{Host: "127.0.0.1", Port: 25}
     err := d.DialAndSend(m)
     fmt.Printf("%v\n",m)
     Checkerr(err) 
}


// mail add image


type Mbox struct{
     From string
     To   string
     Image map[string]string
     Title string
     Msg string
}



func Mailimage(msgbox *Mbox){
     tos:=strings.Split(msgbox.To,",")
     m := gomail.NewMessage()
     m.SetHeader("From", msgbox.From)
     m.SetHeader("To", tos...)
     m.SetHeader("Subject", msgbox.Title)
     for _,ipath:=range msgbox.Image{
         m.Embed(ipath)
     } 
     m.SetBody("text/html", msgbox.Msg)     
     d := gomail.Dialer{Host: "127.0.0.1", Port: 25}
     err := d.DialAndSend(m)
     Checkerr(err) 

}

func Getmaxandmin(arr []float64) (float64,float64){
     maxVal := arr[0]
     minVal :=arr[0]
     for i := 1; i < len(arr); i++ {
        if maxVal < arr[i] {
            maxVal = arr[i]
        }
        if minVal > arr[i]{
           minVal = arr[i]
        }
    }
    return maxVal,minVal
}

