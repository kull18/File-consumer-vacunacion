package utils

import (
    "bytes"
    "log"
    "net/http"
    "time"
)

func SendToAPI(urlApi string, token string, data []byte) error {
    req, err := http.NewRequest("POST", urlApi, bytes.NewBuffer(data))
    if err != nil {
        return err
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", token)

    client := &http.Client{
        Timeout: time.Second * 10, 
    }

    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    log.Println("API response status: ", resp.Status)
    return nil
}