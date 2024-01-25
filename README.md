1. **The 'network' file contains the network topology structure I used, which includes 3 orderer nodes; under the organization org1, there are two peer nodes, peer0 and peer1; under the organization org2, there are also two peer nodes, peer0 and peer1.**



2. **The 'caliper-benchmarks' is the benchmark test I used for the experiment, utilizing the SmallBank workload for testing.**



3. **The 'MyCode' file contains my improvements to the blockchain. The sorting algorithm is implemented in 'orderer-common-blockcutter-scheduler'.**





### To run: 

1. Package the 'MyCode' file into a Docker image using the 'make docker' command to generate a Docker image. 
2. Deploy the network defined in the 'network' file on the server where you want to test. 
3. Use the benchmarks to test the system's performance.
