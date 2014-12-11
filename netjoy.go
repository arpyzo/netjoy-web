package main

import (
	"fmt"
    "net"
    "net/http"
    "html/template"
    "encoding/binary"
    "encoding/json"
    "database/sql"
    _ "github.com/lib/pq"
)

type PacketData struct {
    Ethertype       uint16  `json:"ethertype"`
    SourceIP        string  `json:"source_ip"`
    DestinationIP   string  `json:"destination_ip"`
    Count           uint32  `json:"count"`
    Length          uint32  `json:"length"`
}

func handlerDefault(writer http.ResponseWriter, request *http.Request) {
    fmt.Fprintf(writer, "<body>")
    fmt.Fprintf(writer, "<a href=\"/sqldata?view=ethertype\">SQL Data (ethertype)</a><br>")
    fmt.Fprintf(writer, "<a href=\"/sqldata?view=source_ip\">SQL Data (source ip)</a><br>")
    fmt.Fprintf(writer, "<a href=\"/d3display?view=ethertype\">D3 Display (ethertype)</a><br>")
    fmt.Fprintf(writer, "<a href=\"/d3display?view=source_ip\">D3 Display (source ip)</a><br>")
    fmt.Fprintf(writer, "</body>")
}

func handlerStatic(writer http.ResponseWriter, request *http.Request) {
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
    } else if request.URL.Query().Get("view") == "source_ip" {
        rows, err := db.Query("select source_ip, count(source_ip), sum(packet_length) from netjoy_test group by source_ip")
        if err != nil {
            fmt.Println(err)
        }

        for rows.Next() {
            db_row := PacketData{}
            var int_ip uint32
            err := rows.Scan(&int_ip, &db_row.Count, &db_row.Length)
            if err != nil {
                fmt.Println(err)
            }
            
            db_row.SourceIP = ipStringFromInt(int_ip)           
            db_packet_data = append(db_packet_data, db_row)
        }    
    } else {
        rows, err := db.Query("select destination_ip, count(destination_ip), sum(packet_length) from netjoy_test group by destination_ip")
        if err != nil {
            fmt.Println(err)
        }

        for rows.Next() {
            db_row := PacketData{}
            var int_ip uint32
            err := rows.Scan(&int_ip, &db_row.Count, &db_row.Length)
            if err != nil {
                fmt.Println(err)
            }
            
            db_row.DestinationIP = ipStringFromInt(int_ip)           
            db_packet_data = append(db_packet_data, db_row)
        }    
    }
    
    db.Close()
    
    db_json, err := json.Marshal(db_packet_data)
    if err != nil {
        fmt.Println(err)
    }
    //fmt.Println(string(db_json))
    fmt.Fprintf(writer, "%s", db_json)
}

func handlerD3Display(writer http.ResponseWriter, request *http.Request) {
    testTemplate, _ := template.ParseFiles("d3display.html")
    testTemplate.Execute(writer, nil)
}

func ipStringFromInt(int_ip uint32) string {
    ip := make(net.IP, 4)
    binary.BigEndian.PutUint32(ip, int_ip)
    return ip.String()
}

func main() {  
    http.HandleFunc("/", handlerDefault)
    http.HandleFunc("/images/", handlerStatic)
    http.HandleFunc("/css/", handlerStatic)
    http.HandleFunc("/js/", handlerStatic)
    http.HandleFunc("/sqldata", handlerSqlData)
    http.HandleFunc("/d3display", handlerD3Display)
    http.ListenAndServe("localhost:8080", nil)
}