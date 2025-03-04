# Installation and Usage of parquet-tools

## Table of Contents

- [Installation and Usage of parquet-tools](#installation-and-usage-of-parquet-tools)
  - [Table of Contents](#table-of-contents)
  - [Installation](#installation)
    - [Install from Source](#install-from-source)
    - [Download Pre-built Binaries](#download-pre-built-binaries)
    - [Brew Install](#brew-install)
    - [Container Image](#container-image)
    - [Prebuilt Packages](#prebuilt-packages)
  - [Usage](#usage)
    - [Obtain Help](#obtain-help)
    - [Parquet File Location](#parquet-file-location)
      - [File System](#file-system)
      - [S3 Bucket](#s3-bucket)
      - [GCS Bucket](#gcs-bucket)
      - [Azure Storage Container](#azure-storage-container)
      - [HTTP Endpoint](#http-endpoint)
    - [cat Command](#cat-command)
      - [Full Data Set](#full-data-set)
      - [Skip Rows](#skip-rows)
      - [Limit Number of Rows](#limit-number-of-rows)
      - [Sampling](#sampling)
      - [Compound Rule](#compound-rule)
      - [Output Format](#output-format)
    - [import Command](#import-command)
      - [Import from CSV](#import-from-csv)
      - [Import from JSON](#import-from-json)
      - [Import from JSONL](#import-from-jsonl)
    - [meta Command](#meta-command)
      - [Show Meta Data](#show-meta-data)
      - [Show Meta Data with Base64-encoded Values](#show-meta-data-with-base64-encoded-values)
    - [row-count Command](#row-count-command)
      - [Show Number of Rows](#show-number-of-rows)
    - [schema Command](#schema-command)
      - [JSON Format](#json-format)
      - [Raw Format](#raw-format)
      - [Go Struct Format](#go-struct-format)
    - [shell-completions Command (Experimental)](#shell-completions-command-experimental)
      - [Install Shell Completions](#install-shell-completions)
      - [Uninstall Shell Completions](#uninstall-shell-completions)
      - [Use Shell Completions](#use-shell-completions)
    - [size Command](#size-command)
      - [Show Raw Size](#show-raw-size)
      - [Show Footer Size in JSON Format](#show-footer-size-in-json-format)
      - [Show All Sizes in JSON Format](#show-all-sizes-in-json-format)
    - [version Command](#version-command)
      - [Print Version](#print-version)
      - [Print Version and Build Time in JSON Format](#print-version-and-build-time-in-json-format)
      - [Print Version in JSON Format](#print-version-in-json-format)

## Installation

You can choose one of the installation methods from below, the functionality will be mostly the same.

### Install from Source

Good for people who are familiar with [Go](https://golang.org/), you need 1.17 or newer version.

```bash
$ go get github.com/hangxie/parquet-tools
```

it will install latest stable version of `parquet-tools` to $GOPATH/bin, if you do not set `GOPATH` environment variable explicitly, then its default value can be obtained by running `go evn GOPATH`, usually it is `go/` directory under your home directory.

`parquet-tools` installed from source will not report proper version and build time, so if you run `parquet-tools version`, it will just give you an empty line, all other functions are not affected.

### Download Pre-built Binaries

Good for people do not want to build and all other installation approach do not work.

Go to [relase page](https://github.com/hangxie/parquet-tools/releases), pick the release and platform you want to run, download the corresponding gz/zip file, extract it to your local disk, make sure the execution bit is set if you are running on Linux or Mac, then run the program.

For Windows 10 on ARM (like Surface Pro X), use either windows-arm64 or windows-386 build, if you are in Windows Insider program, windows-amd64 build should work too.

### Brew Install

Mac user can use [Homebrew](https://brew.sh/) to install, it is not part of core formula yet but you can run:

```bash
$ brew uninstall parquet-tools
$ brew tap hangxie/tap
$ brew install go-parquet-tools
```

`parquet-tools` installed by brew is a similar tool built by Java, however, it is [deprecated](https://mvnrepository.com/artifact/org.apache.parquet/parquet-tools-deprecated), since both packages install same `parquet-tools` utility so you need to remove one before installing the other one.

Whenever you want to upgrade to latest version which you should:

```bash
$ brew upgrade go-parquet-tools
```

### Container Image

Container image supports amd64, arm64, and arm/v7, it is hosted in two registries:

* [Docker Hub](https://hub.docker.com/r/hangxie/parquet-tools)
* [GitHub Packages](https://github.com/users/hangxie/packages/container/package/parquet-tools)

You can pull the image from either location:

```bash
$ docker run --rm hangxie/parquet-tools version
v1.13.1
$ podman run --rm ghcr.io/hangxie/parquet-tools version
v1.13.1
```

### Prebuilt Packages

RPM and deb package can be found on [release page](https://github.com/hangxie/parquet-tools/releases), only amd64/x86_64 and arm64/aarch64 arch are available at this moment, download the proper package and run corresponding installation command:

* On Debian/Ubuntu:

```bash
$ sudo dpkg -i  parquet-tools_1.13.0_amd64.deb
Preparing to unpack parquet-tools_1.13.0_amd64.deb ...
Unpacking parquet-tools (1.13.0) ...
Setting up parquet-tools (1.13.0) ...
```

* On CentOS/Fedora:

```bash
$ sudo rpm -Uhv parquet-tools-1.13.0-1.x86_64.rpm
Verifying...                          ################################# [100%]
Preparing...                          ################################# [100%]
Updating / installing...
   1:parquet-tools-1.13.0-1           ################################# [100%]
```

## Usage

### Obtain Help
`parquet-tools` provides help information through `-h` flag, whenever you are not sure about parmater for a command, just add `-h` to the end of the line then it will give you all available options, for example:

```bash
$ parquet-tools meta -h
Usage: parquet-tools meta <uri>

Prints the metadata.

Arguments:
  <uri>    URI of Parquet file, check https://github.com/hangxie/parquet-tools/blob/main/USAGE.md#parquet-file-location for more details.

Flags:
  -h, --help       Show context-sensitive help.

  -b, --base-64    Encode min/max value.
```

Most commands can output JSON format result which can be processed by utilities like [jq](https://stedolan.github.io/jq/) or [JSON parser online](https://jsonparseronline.com/).

### Parquet File Location

`parquet-tools` can read and write parquet files from these locations:
* file system
* AWS Simple Storage Service (S3) bucket
* Google Cloud Storage (GCS) bucket
* Azure Storage Container

`parquet-tools` can read parquet files from these locations:
* HTTP/HTTPS URL

you need to have proper permission on the file you are going to process.

#### File System

For files from file system, you can specify `file://` scheme or just ignore it:

```bash
$ parquet-tools row-count cmd/testdata/good.parquet
4
$ parquet-tools row-count file://cmd/testdata/good.parquet
4
$ parquet-tools row-count file://./cmd/testdata/good.parquet
4
```

#### S3 Bucket

Use full S3 URL to indicate S3 object location, it starts with `s3://`. You need to make sure you have permission to read or write the S3 object, the easiest way to verify that is using [AWS cli](https://aws.amazon.com/cli/):

```bash
$ aws sts get-caller-identity
{
    "UserId": "REDACTED",
    "Account": "123456789012",
    "Arn": "arn:aws:iam::123456789012:user/redacted"
}
$ aws s3 ls s3://aws-roda-hcls-datalake/gnomad/chrm/run-DataSink0-1-part-block-0-r-00000-snappy.parquet
2021-09-08 12:22:56     260887 run-DataSink0-1-part-block-0-r-00000-snappy.parquet
$ parquet-tools row-count s3://aws-roda-hcls-datalake/gnomad/chrm/run-DataSink0-1-part-block-0-r-00000-snappy.parquet
908
```

If an S3 object is publicly accessible and you do not have AWS credential, you can use `--is-public` flag to bypass AWS authentation:

```bash
$ aws sts get-caller-identity

Unable to locate credentials. You can configure credentials by running "aws configure".
$ aws s3 ls --no-sign-request s3://aws-roda-hcls-datalake/gnomad/chrm/run-DataSink0-1-part-block-0-r-00000-snappy.parquet
2021-09-08 12:22:56     260887 run-DataSink0-1-part-block-0-r-00000-snappy.parquet
$ parquet-tools row-count --is-public s3://aws-roda-hcls-datalake/gnomad/chrm/run-DataSink0-1-part-block-0-r-00000-snappy.parquet
908
```

Optionally, you can specify object version by using `--object-version` when you performance read operation (like cat, row-count, schema, etc.) from S3, `parquet-tools` will access current version if this parameter is omitted, if version for the S3 object does not exist, `parquet-tools` will report error:

```bash
$ parquet-tools row-count s3://aws-roda-hcls-datalake/gnomad/chrm/run-DataSink0-1-part-block-0-r-00000-snappy.parquet --object-version non-existent-version
parquet-tools: error: failed to open S3 object [s3://aws-roda-hcls-datalake/gnomad/chrm/run-DataSink0-1-part-block-0-r-00000-snappy.parquet] version [non-existent-version]: operation error S3: HeadObject, https response error StatusCode: 403, RequestID: REDACTED, HostID: REDACTED, api error Forbidden: Forbidden
```

> According to [HeadObject](https://docs.aws.amazon.com/AmazonS3/latest/API/API_HeadObject.html) and [GetObject](https://docs.aws.amazon.com/AmazonS3/latest/API/API_GetObject.html), status code for non-existent object or version will be 403 instead of 404 if the caller does not have permission to `ListBucket`. 

Thanks to [parquet-go-source](https://github.com/xitongsys/parquet-go-source), `parquet-tools` loads only necessary data from S3 bucket, for most cases it is footer only, so it is much more faster than downloading the file from S3 bucket and run `parquet-tools` on a local file. Size of the S3 object used in above sample is more than 4GB, but the `row-count` command takes just several seconds to finish.

#### GCS Bucket

Use full [gsutil](https://cloud.google.com/storage/docs/gsutil) URI to point to GCS object location, it starts with `gs://`. You need to make sure you have permission to read or write to the GSC object, either use application default or GOOGLE_APPLICATION_CREDENTIALS, you can refer to [Google Cloud document](https://cloud.google.com/docs/authentication/production#automatically) for more details.

```bash
$ export GOOGLE_APPLICATION_CREDENTIALS=/path/to/service/account/key.json
$ parquet-tools import -s cmd/testdata/csv.source -m cmd/testdata/csv.schema gs://REDACTED/csv.parquet
$ parquet-tools row-count gs://REDACTED/csv.parquet
7
```

Similar to S3, `parquet-tools` downloads only necessary data from GCS bucket.

#### Azure Storage Container

`parquet-tools` uses the [HDFS URL format](https://docs.microsoft.com/en-us/azure/hdinsight/hdinsight-hadoop-use-blob-storage#access-files-from-within-cluster):
* starts with `wasbs://` (`wasb://` is not supported), followed by
* container as user name, followed by
* storage account as host, followed by
* blob name as path

for example:

> wasbs://public@pandemicdatalake.blob.core.windows.net/curated/covid-19/bing_covid-19_data/latest/bing_covid-19_data.parquet

means the parquet file is at:
* storage account `pandemicdatalake`
* container `public`
* blob `curated/covid-19/bing_covid-19_data/latest/bing_covid-19_data.parquet`

`parquet-tools` uses `AZURE_STORAGE_ACCESS_KEY` environment varialbe to identity access, if the blob is public accessible, then `AZURE_STORAGE_ACCESS_KEY` needs to be either empty or unset to indicate that anonmous access is expected.

```bash
$ AZURE_STORAGE_ACCESS_KEY=REDACTED parquet-tools import -s cmd/testdata/csv.source -m cmd/testdata/csv.schema wasbs://parquet-tools@REDACTED.blob.core.windows.net/test/csv.parquet
$ AZURE_STORAGE_ACCESS_KEY=REDACTED parquet-tools row-count wasbs://public@pandemicdatalake.blob.core.windows.net/curated/covid-19/bing_covid-19_data/latest/bing_covid-19_data.parquet
7
$ AZURE_STORAGE_ACCESS_KEY= parquet-tools row-count wasbs://public@pandemicdatalake.blob.core.windows.net/curated/covid-19/bing_covid-19_data/latest/bing_covid-19_data.parquet
2786653
```

Similar to S3 and GCS, `parquet-tools` downloads only necessary data from blob.

#### HTTP Endpoint

`parquet-tools` supoorts URI with `http` or `https` scheme, the remote server needs to support [Range header](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Range), particularly with unit of `bytes`.

HTTP endpoint does not support write operation so it cannot be used as destination of `import` command.

These options can be used along with HTTP endpoints:
* `--http-multiple-connection` will enable dedicated transport for concurrent requests, `parquet-tools` will establish multiple TCP connections to remote server. This may or may not have performance impact depends on how remote server handles concurrent connections, it is recommended to leave it to default `false` value for all commands except `cat`, and test performance carefully with `cat` command.
* `--http-extra-headers` in the format of `key1=value1,key2=value2,...`, they will be used as extra HTTP headers, a use case is to use them for authentication/authorization that is required by remote server.
* `--http-ignore-tls-error` will igore TLS errors.

```bash
$ parquet-tools row-count https://pandemicdatalake.blob.core.windows.net/public/curated/covid-19/bing_covid-19_data/latest/bing_covid-19_data.parquet
3029995
$ parquet-tools size https://dpla-provider-export.s3.amazonaws.com/2021/04/all.parquet/part-00000-471427c6-8097-428d-9703-a751a6572cca-c000.snappy.parquet
4632041101
```

Similar to S3 and other remote endpoints, `parquet-tools` downloads only necessary data from remote server through [Range header](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Range).

### cat Command

`cat` command output data in parquet file in JSON or JSONL format. Due to most parquet files are rather large, you should use `row-count` command to have a rough idea how many rows are there in the parquet file, then use `--skip`, `--limit` and `--sample-ratio` flags to reduces the output to a certain level, these flags can be used together.

There is a `--page-size` parameter that you probably will never touch it, it tells how many rows `parquet-tools` needs to read from the parquet file every time, you can play with it if you hit performance or resource problem.

#### Full Data Set

```bash
$ parquet-tools cat --format jsonl cmd/testdata/good.parquet
{"Shoe_brand":"shoe_brand","Shoe_name":"shoe_name"}
{"Shoe_brand":"nike","Shoe_name":"air_griffey"}
{"Shoe_brand":"fila","Shoe_name":"grant_hill_2"}
{"Shoe_brand":"steph_curry","Shoe_name":"curry7"}
```

#### Skip Rows

`--skip` is similar to OFFSET in SQL, `parquet-tools` will skip this many rows from beginning of the parquet file before applying other logics.

```bash
$ parquet-tools cat --skip 2 --format jsonl cmd/testdata/good.parquet
{"Shoe_brand":"fila","Shoe_name":"grant_hill_2"}
{"Shoe_brand":"steph_curry","Shoe_name":"curry7"}
```

`parquet-tools` will not report error if `--skip` is greater than total number of rows in parquet file.

```bash
$ parquet-tools cat --skip 20 cmd/testdata/good.parquet
[]
```

#### Limit Number of Rows

`--limit` is similar to LIMIT in SQL, or `head` in Linux shell, `parquet-tools` will stop running after this many rows outputs.

```bash
$ parquet-tools cat --limit 2 cmd/testdata/good.parquet
[{"Shoe_brand":"shoe_brand","Shoe_name":"shoe_name"},{"Shoe_brand":"nike","Shoe_name":"air_griffey"}]
```

#### Sampling

`--sample-ratio` enables sampling, the ration is a number between 0.0 and 1.0 inclusively. `1.0` means output everything in the parquet file, while `0.0` means nothing. If you want to have 1 rows out of every 10 rows, use `0.1`.

This feature picks rows in parquet file randomly, so only `0.0` and `1.0` will output deterministic result, all other ratio may generate data set less or more than you want.

```bash
$ parquet-tools cat --sample-ratio 0.25 cmd/testdata/good.parquet
[{"Shoe_brand":"steph_curry","Shoe_name":"curry7"}]
$ parquet-tools cat --sample-ratio 0.25 cmd/testdata/good.parquet
[{"Shoe_brand":"fila","Shoe_name":"grant_hill_2"}]
$ parquet-tools cat --sample-ratio 0.25 cmd/testdata/good.parquet
[{"Shoe_brand":"shoe_brand","Shoe_name":"shoe_name"}]
$ parquet-tools cat --sample-ratio 0.25 cmd/testdata/good.parquet
[{"Shoe_brand":"fila","Shoe_name":"grant_hill_2"}]
$ parquet-tools cat --sample-ratio 0.25 cmd/testdata/good.parquet
[{"Shoe_brand":"nike","Shoe_name":"air_griffey"},{"Shoe_brand":"fila","Shoe_name":"grant_hill_2"},{"Shoe_brand":"steph_curry","Shoe_name":"curry7"}]
$ parquet-tools cat --sample-ratio 0.25 cmd/testdata/good.parquet
[]
$ parquet-tools cat --sample-ratio 1.0 cmd/testdata/good.parquet
[{"Shoe_brand":"shoe_brand","Shoe_name":"shoe_name"},{"Shoe_brand":"nike","Shoe_name":"air_griffey"},{"Shoe_brand":"fila","Shoe_name":"grant_hill_2"},{"Shoe_brand":"steph_curry","Shoe_name":"curry7"}]
$ parquet-tools cat --sample-ratio 0.0 cmd/testdata/good.parquet
[]
```

#### Compound Rule

`--skip`, `--limit` and `--sample-ratio` can be used together to achieve certain goals, for example, to get the 3rd row from the parquet file:

```bash
$ parquet-tools cat --skip 2 --limit 1 cmd/testdata/good.parquet
{"Shoe_brand":"fila","Shoe_name":"grant_hill_2"}
```

#### Output Format
`cat` supports two output formats, one is the default JSON format that wraps all JSON objects into an array, this works perfectly with small output and is compatible with most JSON toolchains, however, since almost all JSON libraries load full JSON into memory to parse and process, this will lead to memory pressure if you dump a huge amount of data.

```bash
$ parquet-tools cat cmd/testdata/good.parquet
[{"Shoe_brand":"shoe_brand","Shoe_name":"shoe_name"},{"Shoe_brand":"nike","Shoe_name":"air_griffey"},{"Shoe_brand":"fila","Shoe_name":"grant_hill_2"},{"Shoe_brand":"steph_curry","Shoe_name":"curry7"}]
```

`cat` also supports [line delimited JSON streaming format](https://en.wikipedia.org/wiki/JSON_streaming#Line-delimited_JSON_2) format by specifying `--format jsonl`, allows reader of the output to process in a streaming manner, which will greatly reduce the memory footprint. Note that there is always a newline by end of the output.

When you want to filter data, use JSONL format output and pipe to `jq`.

```bash
$ parquet-tools cat --format jsonl cmd/testdata/good.parquet
{"Shoe_brand":"shoe_brand","Shoe_name":"shoe_name"}
{"Shoe_brand":"nike","Shoe_name":"air_griffey"}
{"Shoe_brand":"fila","Shoe_name":"grant_hill_2"}
{"Shoe_brand":"steph_curry","Shoe_name":"curry7"}
```

You can read data line by line and parse every single line as a JSON object if you do not have a toolchain to process JSONL format.

### import Command

`import` command creates a paruet file based from data in other format. The target file can be on local file system or cloud storage object like S3, you need to have permission to write to target location. Existing file or cloud storage object will be overwritten.

The command takes 3 parameters, `--source` tells which file (file system only) to load source data, `--format` tells format of the source data file, it can be `json`, `jsonl` or `csv`, `--schema` points to the file holds schema.

Each source data file format has its own dedicated schema format:

* CSV: you can refer to [sample in this repo](https://github.com/hangxie/parquet-tools/blob/main/cmd/testdata/csv.schema).
* JSON: you can refer to [sample in this repo](https://github.com/hangxie/parquet-tools/blob/main/cmd/testdata/json.schema).
* JSONL: use same schema as JSON format.

#### Import from CSV

```bash
$ parquet-tools import -f csv -s cmd/testdata/csv.source -m cmd/testdata/csv.schema /tmp/csv.parquet
$ parquet-tools row-count /tmp/csv.parquet
7
```

#### Import from JSON

```bash
$ parquet-tools import -f json -s cmd/testdata/json.source -m cmd/testdata/json.schema /tmp/json.parquet
$ parquet-tools row-count /tmp/json.parquet
1
```

As most JSON processing utilities, the whole JSON file needs to be loaded to memory and is treated as single object, so memory footprint may be significant if you try to load a large JSON file. You should use JSONL format if you deal with large amount of data.

#### Import from JSONL

JSONL is [line-delimited JSON streaming format](https://en.wikipedia.org/wiki/JSON_streaming#Line-delimited_JSON), use JSONL if you want to load multiple JSON objects into parquet.

```bash
$ parquet-tools import -f jsonl -s cmd/testdata/jsonl.source -m cmd/testdata/jsonl.schema /tmp/jsonl.parquet
$ parquet-tools row-count /tmp/jsonl.parquet
7
```

### meta Command

`meta` command shows meta data of every row group in a parquet file.

Note that MinValue and MaxValue always show value with base type instead of converted type, i.e. INT32 instead of UINT_8. The `--base64` flag applies to column with type `BYTE_ARRAY` or `FIXED_LEN_BYTE_ARRAY` only, it tells `parquet-tools` to output base64 encoded MinValue and MaxValue of a column, otherwise those values will be shown as UTF8 string.

#### Show Meta Data

```bash
$ parquet-tools meta cmd/testdata/good.parquet
{"NumRowGroups":1,"RowGroups":[{"NumRows":4,"TotalByteSize":349,"Columns":[{"PathInSchema":["Shoe_brand"],"Type":"BYTE_ARRAY","Encodings":["RLE","BIT_PACKED","PLAIN"],"CompressedSize":165,"UncompressedSize":161,"NumValues":4,"NullCount":0,"MaxValue":"steph_curry","MinValue":"fila"},{"PathInSchema":["Shoe_name"],"Type":"BYTE_ARRAY","Encodings":["RLE","BIT_PACKED","PLAIN"],"CompressedSize":192,"UncompressedSize":188,"NumValues":4,"NullCount":0,"MaxValue":"shoe_name","MinValue":"air_griffey"}]}]}
```

#### Show Meta Data with Base64-encoded Values

```bash
$ parquet-tools meta --base64 cmd/testdata/good.parquet
{"NumRowGroups":1,"RowGroups":[{"NumRows":4,"TotalByteSize":349,"Columns":[{"PathInSchema":["Shoe_brand"],"Type":"BYTE_ARRAY","Encodings":["RLE","BIT_PACKED","PLAIN"],"CompressedSize":165,"UncompressedSize":161,"NumValues":4,"NullCount":0,"MaxValue":"c3RlcGhfY3Vycnk=","MinValue":"ZmlsYQ=="},{"PathInSchema":["Shoe_name"],"Type":"BYTE_ARRAY","Encodings":["RLE","BIT_PACKED","PLAIN"],"CompressedSize":192,"UncompressedSize":188,"NumValues":4,"NullCount":0,"MaxValue":"c2hvZV9uYW1l","MinValue":"YWlyX2dyaWZmZXk="}]}]}
```

Note that MinValue, MaxValue and NullCount are optional, if they do not show up in output then it means parquet file does not have that section.

### row-count Command

`row-count` command provides total number of rows in the parquet file:

#### Show Number of Rows

```bash
$ parquet-tools row-count cmd/testdata/good.parquet
4
```

### schema Command

`schema` command shows schema of the parquet file in differnt formats.

#### JSON Format

JSON format schema can be used directly in parquet-go based golang program like [this example](https://github.com/xitongsys/parquet-go/blob/master/example/json_schema.go):

```bash
$ parquet-tools schema cmd/testdata/good.parquet
{"Tag":"name=Parquet_go_root, repetitiontype=REQUIRED","Fields":[{"Tag":"name=Shoe_brand, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=REQUIRED"},{"Tag":"name=Shoe_name, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=REQUIRED"}]}
```

#### Raw Format

Raw format is the schema directly dumped from parquet file, all other formats are derived from raw format.

```bash
$ parquet-tools schema --format raw cmd/testdata/good.parquet
{"repetition_type":"REQUIRED","name":"Parquet_go_root","num_children":2,"children":[{"type":"BYTE_ARRAY","type_length":0,"repetition_type":"REQUIRED","name":"Shoe_brand","converted_type":"UTF8","scale":0,"precision":0,"field_id":0,"logicalType":{"STRING":{}}},{"type":"BYTE_ARRAY","type_length":0,"repetition_type":"REQUIRED","name":"Shoe_name","converted_type":"UTF8","scale":0,"precision":0,"field_id":0,"logicalType":{"STRING":{}}}]}
```

#### Go Struct Format

go struct format generate go struct definition snippet that can be used in go:

```bash
$ parquet-tools schema --format go cmd/testdata/good.parquet | gofmt
type Parquet_go_root struct {
	Shoe_brand string `parquet:"name=Shoe_brand, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=REQUIRED"`
	Shoe_name  string `parquet:"name=Shoe_name, type=BYTE_ARRAY, convertedtype=UTF8, repetitiontype=REQUIRED"`
}
```

based on your use case, type `Parquet_go_root` may need to be renamed.

### shell-completions Command (Experimental)

`shell-completions` updates shell's rcfile with proper shell completions setting, this is an experimental feature at this moment, only bash is tested.

#### Install Shell Completions

To install shell completions. run:

```bash
$ parquet-tools shell-completions
```

You will not get output if everything runs well, you can check shell's rcfile, for example, `.bash_profile` or `.bashrc` for bash, to see what it added.

This command will return error if the same line is in shell's rcfile already.

#### Uninstall Shell Completions

To uninstall shell completions, run:

```bash
$ parquet-tools shell-completions --uninstall
```

You will not get output if everything runs well, you can check shell's rcfile, for example, `.bash_profile` or `.bashrc` for bash, to see what it removed.

This command will return error if the line does not exist in shell rcfile.

#### Use Shell Completions

Hit `<TAB>` key in command line when you need hint or want to auto complete current option.

### size Command

`size` command provides various size information, it can be raw data (compressed) size, uncompressed data size, or footer (meta data) size.

#### Show Raw Size

```bash
$ parquet-tools size cmd/testdata/good.parquet
357
```

#### Show Footer Size in JSON Format

```bash
$ parquet-tools size --query footer --json cmd/testdata/good.parquet
{"Footer":316}
```

#### Show All Sizes in JSON Format

```bash
$ parquet-tools size -q all -j cmd/testdata/good.parquet
{"Raw":357,"Uncompressed":349,"Footer":316}
```

### version Command

`version` command provides version and build information, it will be quite helpful when you are troubleshooting a problem from this tool itself.

#### Print Version

```bash
$ parquet-tools version
v1.13.1
```

#### Print Version and Build Time in JSON Format

```bash
$ parquet-tools version --build-time --json
{"Version":"v1.13.1","BuildTime":"2022-03-15T19:22:19-0700"}
```

#### Print Version in JSON Format

```bash
$ parquet-tools version -j
{"Version":"v1.13.1"}
```
