package main

import (
	"fmt"
    "net/http"
    "html/template"
    "encoding/json"
    "database/sql"
    _ "github.com/lib/pq"
)

type PacketData struct {
    Ethertype       uint16  `json:"ethertype"`
    SourceIP        uint32  `json:"source_ip"`
    DestinationIP   uint32  `json:"destination_ip"`
    Count           uint32  `json:"count"`
    Length          uint32  `json:"length"`
}

func handlerDefault(writer http.ResponseWriter, request *http.Request) {
    fmt.Fprintf(writer, "<body>")
    fmt.Fprintf(writer, "<a href=\"/sqldata\">SQL Data</a><br>")
    fmt.Fprintf(writer, "<a href=\"/d3display\">D3 Display</a><br>")
    fmt.Fprintf(writer, "</body>")
}

func handlerJs(writer http.ResponseWriter, request *http.Request) {
    http.ServeFile(writer, request, request.URL.Path[1:])
}

func handlerSqlData(writer http.ResponseWriter, request *http.Request) {
    db, err := sql.Open("postgres", "user=postgres password=postgres dbname=test sslmode=disable")
    if err != nil {
        fmt.Println(err)
        return
    }
    
    db_packet_data := []PacketData{}
    
    if request.URL.Query().Get("view") == "ethertype" {
        rows, err := db.Query("select ethertype, count(ethertype), sum(packet_length) from netjoy_test group by ethertype")
        if err != nil {
            fmt.Println(err)
        }

        for rows.Next() {
            db_row := PacketData{}
            err := rows.Scan(&db_row.Ethertype, &db_row.Count, &db_row.Length)
            if err != nil {
                fmt.Println(err)
            }
            db_packet_data = append(db_packet_data, db_row)
        }
    } else {
        rows, err := db.Query("select source_ip, count(source_ip), sum(packet_length) from netjoy_test group by source_ip")
        if err != nil {
            fmt.Println(err)
        }

        for rows.Next() {
            db_row := PacketData{}
            err := rows.Scan(&db_row.SourceIP, &db_row.Count, &db_row.Length)
            if err != nil {
                fmt.Println(err)
            }
            db_packet_data = append(db_packet_data, db_row)
        }    
    }
    
    db.Close()
    
    db_json, err := json.Marshal(db_packet_data)
    if err != nil {
        fmt.Println(err)
    }
    fmt.Println(string(db_json))
    fmt.Fprintf(writer, "%s", db_json)
}

func handlerD3Display(writer http.ResponseWriter, request *http.Request) {
    testTemplate, _ := template.ParseFiles("d3display.html")
    testTemplate.Execute(writer, nil)
}

func handlerD3Display2(writer http.ResponseWriter, request *http.Request) {
    testTemplate, _ := template.ParseFiles("d3display2.html")
    testTemplate.Execute(writer, nil)
}

func main() {  
    http.HandleFunc("/", handlerDefault)
    http.HandleFunc("/js/", handlerJs)
    http.HandleFunc("/sqldata", handlerSqlData)
    http.HandleFunc("/d3display", handlerD3Display)
    http.HandleFunc("/d3display2", handlerD3Display2)
    http.ListenAndServe("localhost:8080", nil)
}