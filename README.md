This repository is implement and result of `A Large-Scale Empirical Study on Semantic Versioning in Golang Ecosystem`

We will provide visual web pages in the future to provide the following functions for Go developers:

1. Detect breaking changes
2. Watch downstream client programs for third-party libraries
3. Extract breaking change's usages in client programs

Because of the limit of large files, we only provide code here. For code, dataset, and result, please see [here](https://drive.google.com/drive/folders/1Cf9KITHz5p04xZJCkQQo5BZEP6h4Bov8?usp=sharing)

# Introduction

dataset: store dependency graph that can be imported by `neo4j` and origin repository information

impact: store the code to store pkg info for client programs

result: store experimental results

semver: store the code to detect breaking changes

# Note

we give all breaking changes in `result` dir and not give identifier types of client programs because it is so big.