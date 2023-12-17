# go-worker-credit

POC for test purposes.

Worker consumer kafka topics from the go-fund-transfer service

## diagram

kafka <==(topic.credit)==> go-worker-credit (GROUP-02) (post:/add) ==(REST)==> go-credit(Service.Add)
