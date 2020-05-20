# Moxsoar HTTP API Mocking
## Purpose

Moxsoar is designed to act as a *mock*, or *fake* server, sending arbitary, predefined responses
to requests.

It acts as a middle point between unit testing and integration testing, fulfilling the niche where
neither unit tests nor integration tests are practical.

The original use case for Moxsoar was the development of automation workflows that relied on third party enterprise
services, such as ticketing systems. Because the workflows were either short, developed under strict time pressure,
or developed using tools that do not support it, unit testing these was impractical.

## How it works 

Moxsoar is fundamentally a pool of http servers, each assigned an implementation to mock and configured by a simple
json schema.

Each server responds to http queries as closely as possible to the original (real) implementation of the service.

A simple example is below. Take a look at the following *routes.json* file. This file declares the routes and handling
to be served by a mock http server.

```json
  {
  "routes": [
    {
      "path": "/api/now/table/incident/",
      "methods": [
        {
          "matchregex": "/api/now/table/incident",
          "httpmethod": "POST",
          "responseFile": "table.json",
          "responseCode": 200
        }
      ]
    }
  ]
}
```

Here, we have a single route, matching a particular url and a POST request. It will return whatever is in the 
response file (*responseFile*), in this case, a JSON document, as below:

```json
{
  "result": {
    "active": "true",
    "activity_due": "",
    "additional_assignee_list": "",
    "approval": "not requested",
    "approval_history": ""
  }
}
```

That's all there is to it! 


