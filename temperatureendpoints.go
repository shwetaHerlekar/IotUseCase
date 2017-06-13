package temperatureendpoints

import (
	"time"
	"strconv"
		
	"github.com/GoogleCloudPlatform/go-endpoints/endpoints"
	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

type TemperatureInfo struct {
	SensorID  string           `json:"sensorid"`
	Temp      string          `json:"temperature"`
	Humidity  string          `json:"humidity"`
	Time      string        `json:"time"`
}


type TemperatureAddReq struct {
	SensorID  string       `json:"sensorid" endpoints:"req"`
	Temp      string       `json:"temperature" endpoints:"req"`
	Humidity  string       `json:"humidity" endpoints:"req"`
	Timestamp string       `json:"timestamp" endpoints:"req"`
}

type TemperatureList struct{
	Items []*TemperatureInfo    `json:"items"`
}

type TemperatureListReq struct {
	Limit  int  `json:"limit"`
	Sensor  string  `json:"sensor"`
}

type TemperatureService struct {
	//Placeholder struct to define endpoints
}

func createDataset(client *bigquery.Client, datasetID string) error {
	ctx := context.Background()
	// [START bigquery_create_dataset]
	if err := client.Dataset(datasetID).Create(ctx); err != nil {
		return err
	}
	// [END bigquery_create_dataset]
	return nil
}

func createTable(client *bigquery.Client, datasetID, tableID string) error {
	ctx := context.Background()
	// [START bigquery_create_table]
	schema, err := bigquery.InferSchema(TemperatureInfo{})
	if err != nil {
		return err
	}
	table := client.Dataset(datasetID).Table(tableID)
	if err := table.Create(ctx, schema); err != nil {
		return err
	}
	// [END bigquery_create_table]
	return nil
}



func (ts *TemperatureService) Add(c context.Context, r *TemperatureAddReq) error {
	client, err := bigquery.NewClient(c, "healthcare-12")
	
	createDataset(client, "streamingData")
	createTable(client, "streamingData", "sensorData")
	
	u := client.Dataset("streamingData").Table("sensorData").Uploader()
	
	//temp, _ := strconv.ParseFloat(r.Temp, 32)
	//humidity, _ := strconv.ParseFloat(r.Humidity, 32)
	nanoTimestamp, err := strconv.ParseInt(r.Timestamp, 10, 64)
	if err != nil {
		//log.Fatalln("Could not convert timestamp to the time")
	}
	time1 := time.Unix(0, nanoTimestamp)
	loc, _ := time.LoadLocation("Asia/Kolkata")
	time1 = time1.In(loc)

	e := &TemperatureInfo {
		SensorID : r.SensorID,
		Temp :  r.Temp,
		Humidity : r.Humidity,
		Time : time1.String(),
	}

	err =  u.Put(c, e);
	return err
}

func (ts *TemperatureService) List(c context.Context, r *TemperatureListReq) (*TemperatureList, error) {
	 client, err := bigquery.NewClient(c, "healthcare-12");
	temps := make([]*TemperatureInfo, 0, 10);
        if err != nil {
                return nil, err
        }
		
	/*t1 := time.Now()
	
	stime := time.Date(t1.Year(), t1.Month(), t1.Day(), t1.Hour(), t1.Minute(), 0, 0, t1.Location())
	loc, _ := time.LoadLocation("Asia/Kolkata")
	stime = stime.In(loc)
	log.Infof(c,"%v", stime)
	
	t2 := t1.Add(-2 * time.Minute)
	etime := time.Date(t2.Year(), t2.Month(), t2.Day(), t2.Hour(), t2.Minute(), 0, 0, t2.Location())
	etime = etime.In(loc)
	log.Infof(c,"%v", stime)*/
    query := client.Query("");
	  
	
	  if r.Limit == 13 {
		log.Infof(c,"%v", r.Limit)
		query = client.Query(
                "SELECT "+
                 "*"+
                 "FROM streamingData.sensorData where sensorid ='"+
				 +r.sensor+
				 +"' order by time desc limit 1;")
	  }else{
		log.Infof(c,"%v", r.Limit)
		query = client.Query(
                "SELECT "+
                 "*"+
                 "FROM streamingData.sensorData where sensorid ='"+
				 +r.sensor+
				 +"' order by time desc limit 3;")
	  }
		log.Infof(c,"%v", query)
        // Use standard SQL syntax for queries.
        // See: https://cloud.google.com/bigquery/sql-reference/
        query.QueryConfig.UseStandardSQL = true
		it, err := query.Read(c)
		if err != nil {
                return nil, err
        }
		for {
				e := TemperatureInfo {}
				
				err := it.Next(&e)
				
				if err == iterator.Done {
					break
				}
				temps = append(temps,&e)
			}
		return &TemperatureList{temps}, nil
}

func init() {
	api, err := endpoints.RegisterService(&TemperatureService{},
	                                      "temperatures", "v1",
										  "TemperatureService API", true)
	if err != nil {
		//log.Fatalf("Register service error: %v\n", err)
	}

	list := api.MethodByName("List").Info()
	list.HTTPMethod = "GET"
	list.Path = "temperatures"
	list.Name = "list"
	list.Desc = "List more recent temperature records."

	add := api.MethodByName("Add").Info()
	add.HTTPMethod = "POST"
	add.Path = "temperatures"
	add.Name = "add"
	add.Desc = "Add a new temperature record."

	endpoints.HandleHTTP()
}
