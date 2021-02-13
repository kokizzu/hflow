# Summary
hflow is a simple, command-line, debugging http proxy server 

# Features
hflow supports the following features

* Simple command line interface
* Writes captured traffic to stdout for flexible routing of analysis data
* Decrypts both HTTP and HTTPS traffic
* Provides a HFLOW root CA certificate for which can be imported into client's certficate stores for seamless HTTPS traffic interception
* Automatically decodes gzip and brotli encoded responses
* Supports filtering of captured traffic by URL content and response status
* Allows response output to be truncated at a specified number of bytes

# Installation
To install hflow, run the below from a terminal

```
git clone [this repo] && cd hflow && make install
```

This will clone the repo, compile hflow and then copy the resulting `hflow` binary to `/usr/local/bin`. As this location is typically included in the default `$PATH` variable, hflow should become globally available after the install completes. Uninstalling hflow is simply a matter of deleting the file `/usr/local/bin/hflow`.

Once the install has completed, you can delete the cloned repo, should you wish:

# Usage

## Help
Execute the below to output configuration options

```
hflow -h
``` 

## Quick Start
Execute the below to start the hflow proxy server with the default configuration; capture all traffic: 

```
hflow
```

To route traffic to hflow, configure your HTTP client's proxy address values to `127.0.0.1:[port]` specifying `8080` and `4443` as the `[port]` values for HTTP and HTTPS, respectively.

Use your client to make HTTP requests and note the captured HTTP traffic in your terminal (or wherever `stdout` is redirected).

Hit `[CTRL] + C` to stop hflow 

## Understanding the Output
hflow writes http logs in the format below:

*Requests:*
```
>>> [METHOD] URL

HEADERS

BODY

```

*Responses:*
```
<<< [STATUS CODE] [STATUS] from [ORIGIN REQUEST URL]

HEADERS

BODY

```

Note that hflow diagnostic logs are written to `stderr` and hflow capture data is written to `stdout`. As such, you can redirect these two streams of data to seperate destinations. The example below redirects diagnostic output to a log file and leaves capture data defaulting to `stdout`

```
    hflow 2> ./hflow.log 
```

Log output is set to a low verbosity by default and, other than errors that are likely to be relevant, you should see no diagnostic output after the initial start-up process has completed. If you are experiencing issues that are difficult to diagnose however, increasing the log verbosity may help you resolve them. Logs are not written with a level higher than 3. So the below example outputs the most detailed logs to a log file named `hflow.log`

```
    hflow -v=3 2> ./hflow.log 
```

## Specifying Custom HTTP & HTTPS Ports
To assign non-default ports for hflow to proxy over, execute 

```
hflow -p=[http-port] -ps=[https-port]
``` 

## Filtering Captured Traffic
hflow supports basic filtering commands for specific request URLs and response statuses. These are simple `string contains` style tests and are specified using the `-u="[pattern]"` and `-s="pattern"` flags. 

Note that filters specified for a request URL are also applied to the associated response; so a filter that removes a request also removes its associated response.

*Example of Filtering by Request Domain:*
```
hflow -u="example.com"
```

*Example of Filtering by Request Port:*
```
hflow -u=":443"
```

*Example of Filtering by Response Status Code:*
```
hflow -s="200"
```

*Example of Filtering by Response Status:*
```
hflow -s="OK"
```

*Example of Filtering by a Request Domain and Response Status:*
```
hflow -u="example.com" -s="OK"
```

*Example of Filtering by a Request Querystring:*
```
hflow -u="/?qskey=qsval"
```

## Limiting Response Body Output
It may be preferrable, in the early stages of analysis, to see less detail in the capture. To achieve this, hflow supports capping the number of bytes of the captured response bodies it writes using `-l=[max bytes]`. The example below caps the response body output to 100 bytes:

```
hflow -l=100
```

## Outputting Binary Response Bodies
By default, hflow does not write binary response bodies to its captures. If you require this data, specify the `-b` flag as shown below:

```
hflow -b
```

# Installing the HFLOW Root CA Certificate
To avoid HTTP client warnings relating to the safety of connections to secured domains when proxying HTTPS traffic, you may wish to add the HFLOW Root CA Certificate into your HTTP clients trusted CA certificate collection. Note that this is a potential security risk as the HFLOW Root CA Certificate is freely accessible on the internet. As such, this is undertaken at your own risk and it is advised that you untrust the certificate when not using hflow.

If you wish to proceed, the certficate can be exported in PEM format using the below command. The resulting PEM file can then be loaded directly into your HTTP client's truested CA certificate collection.

```
hflow -ca > ./hflow-ca.pem
```