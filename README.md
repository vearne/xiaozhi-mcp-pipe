# xiaozhi-mcp-pipe
A tool to facilitate Xiaozhi's integration with MCP


## Compile
### 1. Compile it yourself
```
make build
```
### 2) Use the precompiled bin file


[releases](https://github.com/vearne/xiaozhi-mcp-pipe/releases)

## Usage
```
export MCP_ENDPOINT=<mcp_endpoint>
./xiaozhi-mcp-pipe command arg1 arg2 ...
```
For example
### 1) python package
```
export MCP_ENDPOINT=<mcp_endpoint>
./xiaozhi-mcp-pipe uvx vearne_akshare_mcp
```
### 2) nodejs package
```
export MCP_ENDPOINT=<mcp_endpoint>
export SEARXNG_URL=<searxng_endpoint>
./xiaozhi-mcp-pipe npx -y mcp-searxng
```
### 3) other 
```
./xiaozhi-mcp-pipe python calculator.py
```

## Thanks
thanks to [78/mcp-calculator](https://github.com/78/mcp-calculator) for the inspiration and ideas.
