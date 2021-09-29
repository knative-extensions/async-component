**Stress Test Guide:**

Stress testing on the ingress can be performed using [JMeter](https://jmeter.apache.org/download_jmeter.cgi) (Version 5.4.1 at
time of initial testing).
Initial testing can be performed in GUI mode, however if pushing the performance limits it can be run from CLI as well.

**To run via GUI**

In order to stress test open the [PerformanceTest.jmx](./PerformanceTest.jmx) file with JMeter and edit these properties:
- In HTTP Request Defaults, edit `Server Name or IP` field to point to your helloworld application. 
  
Recall that you can find this ip by using `kubectl get kservice helloworld-sleep` as seen in the main README.
  
Note: If you don't see the service make sure you are using the default namespace

```
kubectl get kservice helloworld-sleep
NAME               URL                                                       LATESTCREATED         LATESTREADY           READY   REASON
helloworld-sleep   http://helloworld-sleep.default.111.11.111.111.sslip.io   helloworld-sleep-v1   helloworld-sleep-v1   True    
```

![diagram](./JMeter-images/EditServerName.png)
  

- You may also wish to edit the Graph Results and Summary Report output paths.

![diagram](./JMeter-images/DefaultGraphOutput.png)

After editing these fields, simply run the test plan.  

The given test plan will spin up 100 threads with 10 loops of GET requests to the helloworld application and log the outcome.
The number of threads and loop count can be edited in the top level thread group area of the test plan or editing the JMX file.

![diagram](./JMeter-images/ChangeThreads.png)

**To run via command line:**
```
jmeter -n -t /path/to/async-component/test/JMeter/PerformanceTest.jmx -l /path/to/results.jtl`
```
The output file will contain the raw data.  Graphs and Summary can be viewed via JMeter GUI by simply opening the 
resultant JTL file in either the Graph Results or Summary Report within the test plan thread group.    
