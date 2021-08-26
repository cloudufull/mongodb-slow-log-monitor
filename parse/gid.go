package parse 


import  (
        "fmt"
        "regexp"
        "strings"
        "bytes"
)


func (q *Qt)qread() string{
      rstr:=""
      if q.Index==0{
         if len(q.Allstring)>0{
             rstr=string(q.Allstring[q.Index])
         }
      }else if q.Index>len(q.Allstring)-1{
        rstr="eof"
      }else{
        rstr=string(q.Allstring[q.Index])
      }
      q.Index=q.Index+1
      return rstr
}

func (q *Qt)unread(){
      if q.Index>0{
         q.Index=q.Index-1
      }
    }


func (q *Qt)scanQuoted(tp string) string{
     string_lst:=make([]string,3,10)
     flag:= false 
     for {
         r:=q.qread()
         string_lst=append(string_lst,r)
         if r=="\\"{
            flag=true 
         }
         if tp=="'" && r=="'" {
            if !flag{
                goto ENDLOOP 
               }else{
                flag = false
               }
         }
         if tp==`"` && r==`"` {
            if !flag{
                goto ENDLOOP 
               }else{
                flag= false
               }
         }
         if r=="eof"{
            panic(fmt.Sprintf("missing terminating %s character ",tp))
         }
     }
ENDLOOP:
     return strings.Join(string_lst,"")
}


func (q *Qt)findcomma() string{
     string_lst:=make([]string,3,10)
     for {
         r:=q.qread()
         if r=="{" || r=="["{return r}
         string_lst=append(string_lst,r)
         if q.Debug{fmt.Println("~~~~~:",r,"lstat:",q.Lstat)}
         switch r{
                case "'":
                     v:=q.scanQuoted("'")
                     string_lst=append(string_lst,v)
                case `"`:
                     v:=q.scanQuoted(`"`)
                     string_lst=append(string_lst,v)
                case ",":
                     q.unread()
                     goto ENDLOOP 
                case "}": 
                     if q.Stat==1{
                        q.Stat=0
                        q.Indic=0
                        q.unread()
                        goto ENDLOOP
                        }
                case "]":
                     if q.Lstat==1{
                        q.Lstat=0
                        q.unread()
                        goto ENDLOOP}
                case "eof":
                     return strings.Join(string_lst,"") 
                     
         }

    }
ENDLOOP: 
    return strings.Join(string_lst,"")
}


type Qt struct{
     Allstring string
     Index int
     Qstring bytes.Buffer 
     Stat int 
     Lstat int
     Indic int
     Debug bool
}

func (q *Qt)Scan() {
     x,_:=regexp.Compile(`\s`)
     q.Allstring=x.ReplaceAllString(q.Allstring,"")
     for {
         i:=q.qread()
         switch i {
               case "eof":
                     return  
               case `'`:
                  value:=q.scanQuoted(`'`)
                  if q.Lstat==1 && q.Indic==0{
                      q.Qstring.Write([]byte(`?`))
                  }else if q.Stat==0{
                       q.Qstring.Write([]byte(i))
                       q.Qstring.Write([]byte(value))
                  }else{
                       q.Qstring.Write([]byte(`?`))
                       if q.Stat==1{q.Stat=0}
                  }
               case `"`:
                  value:=q.scanQuoted(`"`)
                  if q.Lstat==1 && q.Indic==0{
                     q.Qstring.Write([]byte(`?`))
                  }else if q.Stat==0{
                       q.Qstring.Write([]byte(i))
                       q.Qstring.Write([]byte(value))
                  }else{
                       q.Qstring.Write([]byte(`?`))
                       q.Stat=0
                  }
               case `:`:
                  q.Qstring.Write([]byte(i))
                  q.Stat=1
               case  `{`:
                  q.Qstring.Write([]byte(i))
                  q.Stat=0
                  q.Indic=1
                  if q.Debug{fmt.Println(string(q.Qstring.Bytes()),"---xx---"," q.stat =>",q.Stat,"lstat:",q.Lstat)}
               case `[`:
                  q.Qstring.Write([]byte(i))
                  q.Lstat=1
                  q.Stat=0
                  q.Indic=0
                  if q.Debug{fmt.Println(string(q.Qstring.Bytes()),"--------",q.Lstat)}
               case `}`:
                   q.Indic=0
                   q.Qstring.Write([]byte(i))
               case `]`:
                   q.Lstat=0
                   q.Qstring.Write([]byte(i))
               default: 
                  if v,_:=regexp.MatchString(`\s`,i);!v{
                     if q.Debug{fmt.Println(string(q.Qstring.Bytes()),"lstat:",q.Lstat," stat",q.Stat,"indic:",q.Indic,i)}
                     //q.unread()
                     if q.Stat==1{
                        if q.Debug{fmt.Println("before stat:",q.Stat,q.Lstat)}
                        _=q.findcomma() 
                        q.Qstring.Write([]byte(`?`))
                        q.Stat=0
                     }else if q.Lstat==1 && q.Stat==0 && q.Indic==0{
                        q.Qstring.Write([]byte(i))
                        if q.Debug{fmt.Println("before stat:",q.Stat,q.Lstat,q.Indic)}
                        value:=q.findcomma() 
                        if q.Debug{fmt.Println("=======>",value)}
                        if value=="{" || value=="["{
                            q.Stat=0
                            q.Indic=1
                           q.Qstring.Write([]byte(value))
                        }else{
                            q.Qstring.Write([]byte(`?`))
                        }
                     }else{
                        q.Qstring.Write([]byte(i))
                     }
                 }
           }
    }

}
