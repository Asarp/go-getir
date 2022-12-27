package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Used both for in-memory requests and responses
type MemoryStruct struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Used for DB requests
type DbRequest struct {
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
	MinCount  int    `json:"minCount"`
	MaxCount  int    `json:"maxCount"`
}

// Handles redis-related operations
func handleMemory(w http.ResponseWriter, req *http.Request) {

	// Gets the redis client which will be used in future operations
	rdb := getRedisClient()
	// Setting response headers
	w.Header().Set("Content-Type", "application/json")

	// Empty map object that will act as the body of the response
	resp := make(map[string]string)

	if req.Method == "GET" {
		// Gets the values that are passed under the query variable "key"
		query := req.URL.Query()
		values := query["key"]
		if len(values) == 0 {
			// If there aren't any values, return error
			w.WriteHeader(http.StatusBadRequest)
			resp["message"] = "Invalid Query Variable"
			jsonResp, err := json.Marshal(resp)
			if err != nil {
				// If there is an error during the JSON marshal which is unexpected due to the compatible format of the "resp" map,
				// I wanted the api to still be able to return the Status portion of the response.
				// The "resp" map that corresponds to the response's body is left empty due to the error encountered during marshalling.
				// And error is printed on the console. This behavior of error handling for this specific case is repeated in many places in this project
				fmt.Printf("Error happened in JSON marshal. Err: %s", err)
			}
			w.Write(jsonResp)

		} else {
			// queries redis with the last value passed for "key" (to handle the case of multiple values passed under "key"), returns the result
			r_response, err := fetchMemory(rdb, values[len(values)-1])
			if err != nil {
				// If that last value doesn't have any corresponding element in redis, return error
				w.WriteHeader(http.StatusBadRequest)
				resp["message"] = "Key is not Present"
				jsonResp, err := json.Marshal(resp)
				if err != nil {
					fmt.Printf("Error happened in JSON marshal. Err: %s", err)
				}
				w.Write(jsonResp)
				return
			}
			// If it does, return (key, value) in response
			w.WriteHeader(http.StatusOK)
			jsonResp, err := json.Marshal(r_response)
			if err != nil {
				fmt.Printf("Error happened in JSON marshal. Err: %s", err)
			}
			w.Write(jsonResp)
		}

	} else if req.Method == "POST" {

		// Decodes the request body to an instance of MemoryStruct
		decoder := json.NewDecoder(req.Body)
		var r MemoryStruct
		err := decoder.Decode(&r)

		if err != nil {
			// If request body is in an invalid format, returns error
			w.WriteHeader(http.StatusBadRequest)
			resp["message"] = "Invalid Request"
			jsonResp, err := json.Marshal(resp)
			if err != nil {
				fmt.Printf("Error happened in JSON marshal. Err: %s", err)
			}
			w.Write(jsonResp)
			return
		}
		if err := storeMemory(rdb, r.Key, r.Value); err != nil {
			// If there is an error during storeMemory operation, returns error
			w.WriteHeader(http.StatusBadRequest)
			resp["message"] = "Error During Write"
			jsonResp, err := json.Marshal(resp)
			if err != nil {
				fmt.Printf("Error happened in JSON marshal. Err: %s", err)
			}
			w.Write(jsonResp)
			return
		}
		// If no errors are present, return a success response with {key, value} struct as its body
		w.WriteHeader(http.StatusOK)
		jsonResp, err := json.Marshal(r)
		if err != nil {
			fmt.Printf("Error happened in JSON marshal. Err: %s", err)
		}
		w.Write(jsonResp)

	} else {
		// Request methods other than GET and POST are not allowed
		w.WriteHeader(http.StatusMethodNotAllowed)
		resp["message"] = "Request Method not Allowed"
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			fmt.Printf("Error happened in JSON marshal. Err: %s", err)
		}
		w.Write(jsonResp)
	}
}

func handleDB(w http.ResponseWriter, req *http.Request) {

	// Setting response headers
	w.Header().Set("Content-Type", "application/json")
	// Date format layout that will be used to check the format of start/endDate fields of our request
	const YYYYMMDD = "2006-01-02"
	// Empty DbResponse that will serve as the response body for errors later
	resp := &DbResponse{}

	if req.Method == "GET" {
		// Decodes the request body to an instance of DbRequest
		decoder := json.NewDecoder(req.Body)
		var r DbRequest
		err := decoder.Decode(&r)

		if err != nil {
			// If request body is in an invalid format, returns error
			w.WriteHeader(http.StatusBadRequest)
			resp.Code = -1
			resp.Msg = "Invalid Request Format"
			jsonResp, err := json.Marshal(resp)
			if err != nil {
				fmt.Printf("Error happened in JSON marshal. Err: %s", err)
			}
			w.Write(jsonResp)
			return
		}
		// Parses start/endDates into "YYYY-MM-DD" format
		tStart, err1 := time.Parse(YYYYMMDD, r.StartDate)
		tEnd, err2 := time.Parse(YYYYMMDD, r.EndDate)

		if err1 != nil || err2 != nil {
			// If start/endDate can't be parsed into "YYYY-MM-DD" format, returns error
			w.WriteHeader(http.StatusBadRequest)
			resp.Code = -1
			resp.Msg = "Invalid Time Format"
			jsonResp, err := json.Marshal(resp)
			if err != nil {
				fmt.Printf("Error happened in JSON marshal. Err: %s", err)
			}
			w.Write(jsonResp)
			return
		}

		// Querying the DB
		dbResp, err := fetchDB(tStart, tEnd, r.MinCount, r.MaxCount)

		if err != nil {
			// If there is an error during the fetchDB operation, returns error
			w.WriteHeader(http.StatusBadRequest)
			resp.Code = -1
			resp.Msg = err.Error()
			jsonResp, err := json.Marshal(resp)
			if err != nil {
				fmt.Printf("Error happened in JSON marshal. Err: %s", err)
			}
			w.Write(jsonResp)
			return
		}
		// If fetchDB operation succeeds, return a Success response with the return value from the fetchDb operation as its body
		w.WriteHeader(http.StatusOK)
		jsonResp, err := json.Marshal(dbResp)
		if err != nil {
			fmt.Printf("Error happened in JSON marshal. Err: %s", err)
		}
		w.Write(jsonResp)

	} else {
		// Request methods other than GET are not allowed
		w.WriteHeader(http.StatusMethodNotAllowed)
		resp.Code = -2
		resp.Msg = "Request not Allowed"
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			fmt.Printf("Error happened in JSON marshal when writing response. Err: %s", err)
		}
		w.Write(jsonResp)
	}
}

func server() {
	// Assings handler functions to api endpoints
	http.HandleFunc("/db", handleDB)
	http.HandleFunc("/memory", handleMemory)

	// Starts to listen in in the requests on the port 8080
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
