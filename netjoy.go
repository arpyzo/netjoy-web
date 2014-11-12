package main

import (
	"fmt"
    "strconv"
    "net/http"
    "html/template"
    "encoding/json"
    "database/sql"
    _ "github.com/lib/pq"
)

type PacketData struct {
    Ethertype   uint16  `json:"ethertype"`
    Count       uint32  `json:"count"`
    Length      uint32  `json:"length"`
}

var myVar int

func handlerJs(writer http.ResponseWriter, request *http.Request) {
    http.ServeFile(writer, request, request.URL.Path[1:])
}

func handlerDefault(writer http.ResponseWriter, request *http.Request) {
    fmt.Fprintf(writer, "<body>Default: %s<br>", request.URL.Path)
    fmt.Fprintf(writer, "Remote: %s<br>", request.RemoteAddr)
    fmt.Fprintf(writer, "<a href=\"/special/special_text\">Special</a><br>")
    fmt.Fprintf(writer, "<a href=\"/template\">Template</a><br>")
    fmt.Fprintf(writer, "<a href=\"/sqldata\">SQL Data</a><br>")
    fmt.Fprintf(writer, "<a href=\"/d3data\">D3 Data</a><br>")
    fmt.Fprintf(writer, "</body>")
}

func handlerSpecial(writer http.ResponseWriter, request *http.Request) {
    fmt.Fprintf(writer, "<body>Special: %s<br>", request.URL.Path)
    fmt.Fprintf(writer, "Remote: %s</body>", request.RemoteAddr)
}

func handlerTemplate(writer http.ResponseWriter, request *http.Request) {
    testTemplate, _ := template.ParseFiles("template.html")
    testTemplate.Execute(writer, myVar)
}

func handlerChangeVar(writer http.ResponseWriter, request *http.Request) {
    myVar, _ = strconv.Atoi(request.FormValue("myVar"))
    http.Redirect(writer, request, "/template/", http.StatusFound)
}

func handlerSqlData(writer http.ResponseWriter, request *http.Request) {
    db, err := sql.Open("postgres", "user=postgres password=postgres dbname=test sslmode=disable")
    if err != nil {
        fmt.Println(err)
        return
    }
    
    rows, err := db.Query("select packet_type, count(packet_type), sum(packet_length) from netjoy_test group by packet_type")
    if err != nil {
        fmt.Println(err)
    }

    db_packet_data := []PacketData{}
    for rows.Next() {
        db_row := PacketData{}
        err := rows.Scan(&db_row.Ethertype, &db_row.Count, &db_row.Length)
        if err != nil {
            fmt.Println(err)
        }
        //fmt.Printf("%d --- %d --- %d\n", db_row.ethertype, db_row.count, db_row.length)
        db_packet_data = append(db_packet_data, db_row)
    }

    db.Close()
    
    //fmt.Println(db_packet_data[0])
    //fmt.Println(db_packet_data[1])
    //fmt.Println(db_packet_data[2])
    
    db_json, err := json.Marshal(db_packet_data)
    if err != nil {
        fmt.Println(err)
    }
    fmt.Println(string(db_json))
    fmt.Fprintf(writer, "%s", db_json)
}

func handlerD3Data(writer http.ResponseWriter, request *http.Request) {
    testTemplate, _ := template.ParseFiles("d3data.html")
    testTemplate.Execute(writer, myVar)
}

func main() {
    myVar = 13
    
    http.HandleFunc("/js/", handlerJs)
    http.HandleFunc("/", handlerDefault)
    http.HandleFunc("/special/", handlerSpecial)
    http.HandleFunc("/template", handlerTemplate)
    http.HandleFunc("/change", handlerChangeVar)
    http.HandleFunc("/sqldata", handlerSqlData)
    http.HandleFunc("/d3data", handlerD3Data)
    http.ListenAndServe("localhost:8080", nil)
}