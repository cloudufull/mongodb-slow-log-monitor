package parse


import (
     "database/sql"

)

type DB_save struct{
     db *sql.DB
     Mail *Mail_info
     Debug bool
}



type Save_alert_st struct{
        qtm int 
        atm int
        etime int
        qid string
        db string
        table string
        host string
        optp string
        qmessage string

}


type Mail_info struct{
     Mail_from string
     Mail_to string
     DC string
     Mailbox map[string][]string
}


type Dataset struct{
     Len int
     TT string //title
     X  string //x Label name
     Y  string //y Label name
     Xdt []float64
     Ydt []float64
     Imagepath string
}
