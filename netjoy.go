package main

import (
	"fmt"
    "net"
    "net/http"
    "net/url"
    "html/template"
    "encoding/json"
    "database/sql"
    _ "github.com/lib/pq"
)

const debug = 1

type PacketData struct {
    Ethertype       uint16  `json:"ethertype"`
    SourceIp        string  `json:"source_ip"`
    DestinationIp   string  `json:"destination_ip"`
    Count           uint32  `json:"count"`
    Length          uint32  `json:"length"`
}

type Parameters struct {
    groupBy         string
    aggregateBy     string
    sourceCidr      string
    destinationCidr string
}

func handlerDefault(writer http.ResponseWriter, request *http.Request) {
    fmt.Fprintf(writer, "<body>")
    fmt.Fprintf(writer, "<a href=\"/sqldata?group_by=ethertype\">SQL Data (ethertype)</a><br>")
    fmt.Fprintf(writer, "<a href=\"/sqldata?group_by=source_ip\">SQL Data (source ip)</a><br>")
    fmt.Fprintf(writer, "<a href=\"/d3display?group_by=ethertype\">D3 Display (ethertype)</a><br>")
    fmt.Fprintf(writer, "<a href=\"/d3display?group_by=source_ip\">D3 Display (source ip)</a><br>")
    fmt.Fprintf(writer, "</body>")
}

func handlerStatic(writer http.ResponseWriter, request *http.Request) {
    http.ServeFile(writer, request, request.URL.Path[1:])
}

func handlerD3Display(writer http.ResponseWriter, request *http.Request) {
    testTemplate, _ := template.ParseFiles("d3display.html")
    testTemplate.Execute(writer, nil)
}

func handlerSqlData(writer http.ResponseWriter, request *http.Request) {
    db, err := sql.Open("postgres", "user=postgres password=postgres dbname=test sslmode=disable")
    if err != nil {
        fmt.Println(err)
        return
    }
    
    if (debug >= 1) {
        fmt.Println("DEBUG - Raw query string: " + request.URL.RawQuery)
    }
    parameters := extractParameters(request.URL.Query())
    if (debug >= 1) {
        fmt.Println("DEBUG - groupBy: " + parameters.groupBy + ", aggregateBy: " + parameters.aggregateBy)
    }

    databaseQuery := createDatabaseQuery(parameters)
    if (debug >= 1) {
        fmt.Println("DEBUG - databaseQuery: " + databaseQuery)
    }
        
    dbPacketData := []PacketData{}
    dbRow := PacketData{}
    
    rows, err := db.Query(databaseQuery)
    if err != nil {
        fmt.Println(err)
        db.Close()
        return
    }
    
    // TODO: figure out how to elimiate repetative code here
    
    switch {
        case parameters.groupBy == "ethertype" && parameters.aggregateBy == "count":
            for rows.Next() {
                err := rows.Scan(&dbRow.Ethertype, &dbRow.Count)
                if err != nil {
                    fmt.Println(err)
                }
                dbPacketData = append(dbPacketData, dbRow)
            }
        case parameters.groupBy == "ethertype" && parameters.aggregateBy == "length":
            for rows.Next() {
                err := rows.Scan(&dbRow.Ethertype, &dbRow.Length)
                if err != nil {
                    fmt.Println(err)
                }
                
                dbPacketData = append(dbPacketData, dbRow)
            }    
        case parameters.groupBy == "source_ip" && parameters.aggregateBy == "count":
            for rows.Next() {
                err := rows.Scan(&dbRow.SourceIp, &dbRow.Count)
                if err != nil {
                    fmt.Println(err)
                }
                dbPacketData = append(dbPacketData, dbRow)
            }
        case parameters.groupBy == "source_ip" && parameters.aggregateBy == "length":
            for rows.Next() {
                err := rows.Scan(&dbRow.SourceIp, &dbRow.Length)
                if err != nil {
                    fmt.Println(err)
                }
                
                dbPacketData = append(dbPacketData, dbRow)
            }    
        case parameters.groupBy == "destination_ip" && parameters.aggregateBy == "count":
        for rows.Next() {
            err := rows.Scan(&dbRow.DestinationIp, &dbRow.Count)
            if err != nil {
                fmt.Println(err)
            }
            dbPacketData = append(dbPacketData, dbRow)
        }
        case parameters.groupBy == "destination_ip" && parameters.aggregateBy == "length":
            for rows.Next() {
                err := rows.Scan(&dbRow.DestinationIp, &dbRow.Length)
                if err != nil {
                    fmt.Println(err)
                }
                
                dbPacketData = append(dbPacketData, dbRow)
            }    
    }
    
    db.Close()
    
    dbJson, err := json.Marshal(dbPacketData)
    if err != nil {
        fmt.Println(err)
        return
    }
    if (debug >= 2) {
        fmt.Println("DEBUG - " + string(dbJson))
    }
    fmt.Fprintf(writer, "%s", dbJson)
}

func extractParameters(queryValues url.Values) Parameters {
    var parameters Parameters
    
    parameters.groupBy = "ethertype"
    groupByList := []string{"ethertype", "source_ip", "destination_ip", "source_port", "destination_port"}
    if (isParameterOk(&groupByList, queryValues.Get("group_by"))) {
        parameters.groupBy  = queryValues.Get("group_by")
    }
    
    parameters.aggregateBy = "length"
    if (queryValues.Get("aggregate_by") == "count") {
        parameters.aggregateBy = "count"
    }
    
    _, _, err := net.ParseCIDR(queryValues.Get("source"))
    if (err == nil) {
        parameters.sourceCidr = queryValues.Get("source")
    }
    
    _, _, err = net.ParseCIDR(queryValues.Get("destination"))
    if (err == nil) {
        parameters.destinationCidr = queryValues.Get("destination")
    }
    
    return parameters
}

func createDatabaseQuery(parameters Parameters) string {
    var databaseQuery string

    aggregateFunc := "sum(packet_length)"
    if (parameters.aggregateBy == "count") {
        aggregateFunc = "count(" + parameters.groupBy + ")"
    }
    
    databaseQuery = "select " + parameters.groupBy + ", " + aggregateFunc
    databaseQuery += " from netjoy_test "
    
    // TODO: Clean up this ugly logic
    if (parameters.sourceCidr != "") {
        databaseQuery += "where '" + parameters.sourceCidr + "' >> source_ip "
        if (parameters.destinationCidr != "") {
            databaseQuery += "and '" + parameters.destinationCidr + "' >> destination_ip "        
        }
    } else if (parameters.destinationCidr != "") {
        databaseQuery += "where '" + parameters.destinationCidr + "' >> destination_ip "
    }
    
    databaseQuery += "group by " + parameters.groupBy + " order by " + aggregateFunc + " desc"
    
    return databaseQuery
}

func isParameterOk(list *[]string, parameter string) bool {
    for _, x := range *list {
        if (parameter == x) {
            return true
        }
    }
    
    return false
}

func main() {  
    http.HandleFunc("/", handlerDefault)
    http.HandleFunc("/images/", handlerStatic)
    http.HandleFunc("/css/", handlerStatic)
    http.HandleFunc("/js/", handlerStatic)
    http.HandleFunc("/d3display", handlerD3Display)
    http.HandleFunc("/sqldata", handlerSqlData)
    http.ListenAndServe("localhost:8080", nil)
}