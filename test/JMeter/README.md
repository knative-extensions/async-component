**Stress Test Guide:**

Stress testing on the ingress was performed using [JMeter](https://jmeter.apache.org/download_jmeter.cgi).
Initial testing was performed in GUI mode, however if pushing the limits it can be run from CLI as well.

In order to stress test open the `PerformanceTest.jmx` file with JMeter and edit these properties:
- In HTTP Request Defaults, edit `Server Name or IP` field to point to your helloworld application.
    Recall that you can find this by using `kubectl get kservice helloworld-sleep
  `
  
- You may also wish to edit the Graph Results and Summary Report output paths.

After editing these fields simply run the test plan.  

The test plan will spin up 100 threads with 10 loops of GET requests to the helloworld application and log the outcome.