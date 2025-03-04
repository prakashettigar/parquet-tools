package cmd

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	pqtazblob "github.com/xitongsys/parquet-go-source/azblob"
	"github.com/xitongsys/parquet-go-source/gcs"
	"github.com/xitongsys/parquet-go-source/http"
	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go-source/s3v2"
	"github.com/xitongsys/parquet-go/parquet"
	"github.com/xitongsys/parquet-go/reader"
	"github.com/xitongsys/parquet-go/source"
	"github.com/xitongsys/parquet-go/types"
	"github.com/xitongsys/parquet-go/writer"
)

const (
	schemeLocal              string = "file"
	schemeAWSS3              string = "s3"
	schemeGoogleCloudStorage string = "gs"
	schemeAzureStorageBlob   string = "wasbs"
	schemeHTTP               string = "http"
	schemeHTTPS              string = "https"
)

// CommonOption represents common options across most commands
type CommonOption struct {
	URI string `arg:"" predictor:"file" help:"URI of Parquet file."`
}

// ReadOption include options for read operation
type ReadOption struct {
	CommonOption
	HTTPMultipleConnection bool              `help:"(HTTP URI only) use multiple HTTP connection." default:"false"`
	HTTPIgnoreTLSError     bool              `help:"(HTTP URI only) ignore TLS error." default:"false"`
	HTTPExtraHeaders       map[string]string `mapsep:"," help:"(HTTP URI only) extra HTTP headers." default:""`
	ObjectVersion          string            `help:"(S3 URI only) object version." default:""`
	IsPublic               bool              `help:"(S3 URI only) object is publicly accessible." default:"false"`
}

// Context represents command's context
type Context struct {
	Version string
	Build   string
}

// ReinterpretField represents a field that needs to be re-interpretted before output
type ReinterpretField struct {
	parquetType   parquet.Type
	convertedType parquet.ConvertedType
	precision     int
	scale         int
}

func parseURI(uri string) (*url.URL, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("unable to parse file location [%s]: %s", uri, err.Error())
	}

	if u.Scheme == "" {
		u.Scheme = schemeLocal
	}

	if u.Scheme == schemeLocal {
		u.Path = filepath.Join(u.Host, u.Path)
		u.Host = ""
	}

	return u, nil
}

func getS3Client(bucket string, isPublic bool) (*s3.Client, error) {
	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(ctx, config.WithDefaultRegion("us-east-1"))
	if err != nil {
		return nil, fmt.Errorf("failed to load config to determine bucket region: %s", err.Error())
	}
	region, err := manager.GetBucketRegion(ctx, s3.NewFromConfig(cfg), bucket)
	if err != nil {
		var apiErr manager.BucketNotFound
		if errors.As(err, &apiErr) {
			return nil, fmt.Errorf("unable to find region of bucket [%s]", bucket)
		}
		return nil, fmt.Errorf("AWS error: %s", err.Error())
	}

	if isPublic {
		return s3.NewFromConfig(aws.Config{Region: region}), nil
	}
	cfg.Region = region
	return s3.NewFromConfig(cfg), nil
}

func newParquetFileReader(option ReadOption) (*reader.ParquetReader, error) {
	u, err := parseURI(option.URI)
	if err != nil {
		return nil, err
	}

	var fileReader source.ParquetFile
	switch u.Scheme {
	case schemeAWSS3:
		s3Client, err := getS3Client(u.Host, option.IsPublic)
		if err != nil {
			return nil, err
		}

		var objVersion *string = nil
		if option.ObjectVersion != "" {
			objVersion = &option.ObjectVersion
		}
		fileReader, err = s3v2.NewS3FileReaderWithClientVersioned(context.Background(), s3Client, u.Host, strings.TrimLeft(u.Path, "/"), objVersion)
		if err != nil {
			return nil, fmt.Errorf("failed to open S3 object [%s] version [%s]: %s", option.URI, option.ObjectVersion, err.Error())
		}
	case schemeLocal:
		fileReader, err = local.NewLocalFileReader(u.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to open local file [%s]: %s", u.Path, err.Error())
		}
	case schemeGoogleCloudStorage:
		fileReader, err = gcs.NewGcsFileReader(context.Background(), "", u.Host, strings.TrimLeft(u.Path, "/"))
		if err != nil {
			return nil, fmt.Errorf("failed to open GCS object [%s]: %s", option.URI, err.Error())
		}
	case schemeAzureStorageBlob:
		azURL, cred, err := azureAccessDetail(*u)
		if err != nil {
			return nil, err
		}

		fileReader, err = pqtazblob.NewAzBlobFileReader(context.Background(), azURL, cred, pqtazblob.ReaderOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to open Azure blob object [%s]: %s", option.URI, err.Error())
		}
	case schemeHTTP, schemeHTTPS:
		fileReader, err = http.NewHttpReader(option.URI, option.HTTPMultipleConnection, option.HTTPIgnoreTLSError, option.HTTPExtraHeaders)
		if err != nil {
			return nil, fmt.Errorf("failed to open HTTP source [%s]: %s", option.URI, err.Error())
		}
	default:
		return nil, fmt.Errorf("unknown location scheme [%s]", u.Scheme)
	}

	return reader.NewParquetReader(fileReader, nil, int64(runtime.NumCPU()))
}

func newFileWriter(option CommonOption) (source.ParquetFile, error) {
	u, err := parseURI(option.URI)
	if err != nil {
		return nil, err
	}

	var fileWriter source.ParquetFile
	switch u.Scheme {
	case schemeAWSS3:
		s3Client, err := getS3Client(u.Host, false)
		if err != nil {
			return nil, err
		}

		fileWriter, err = s3v2.NewS3FileWriterWithClient(context.Background(), s3Client, u.Host, strings.TrimLeft(u.Path, "/"), nil)
		if err != nil {
			return nil, fmt.Errorf("failed to open S3 object [%s]: %s", option.URI, err.Error())
		}
	case schemeLocal:
		fileWriter, err = local.NewLocalFileWriter(u.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to open local file [%s]: %s", u.Path, err.Error())
		}
	case schemeGoogleCloudStorage:
		fileWriter, err = gcs.NewGcsFileWriter(context.Background(), "", u.Host, strings.TrimLeft(u.Path, "/"))
		if err != nil {
			return nil, fmt.Errorf("failed to open GCS object [%s]: %s", option.URI, err.Error())
		}
	case schemeAzureStorageBlob:
		azURL, cred, err := azureAccessDetail(*u)
		if err != nil {
			return nil, err
		}

		fileWriter, err = pqtazblob.NewAzBlobFileWriter(context.Background(), azURL, cred, pqtazblob.WriterOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to open Azure blob object [%s]: %s", option.URI, err.Error())
		}
	case schemeHTTP, schemeHTTPS:
		return nil, fmt.Errorf("writing to %s endpoint is not currently supported", u.Scheme)
	default:
		return nil, fmt.Errorf("unknown location scheme [%s]", u.Scheme)
	}

	return fileWriter, nil
}

func newCSVWriter(option CommonOption, schema []string) (*writer.CSVWriter, error) {
	fileWriter, err := newFileWriter(option)
	if err != nil {
		return nil, err
	}

	return writer.NewCSVWriter(schema, fileWriter, int64(runtime.NumCPU()))
}

func newJSONWriter(option CommonOption, schema string) (*writer.JSONWriter, error) {
	fileWriter, err := newFileWriter(option)
	if err != nil {
		return nil, err
	}

	return writer.NewJSONWriter(schema, fileWriter, int64(runtime.NumCPU()))
}

func azureAccessDetail(azURL url.URL) (string, azblob.Credential, error) {
	container := azURL.User.Username()
	if azURL.Host == "" || container == "" || strings.HasSuffix(azURL.Path, "/") {
		return "", nil, fmt.Errorf("azure blob URI format: wasbs://container@storageaccount.blob.windows.core.net/path/to/blob")
	}
	httpURL := fmt.Sprintf("https://%s/%s%s", azURL.Host, container, azURL.Path)

	accessKey := os.Getenv("AZURE_STORAGE_ACCESS_KEY")
	if accessKey == "" {
		// anonymouse access
		return httpURL, azblob.NewAnonymousCredential(), nil
	}

	credential, err := azblob.NewSharedKeyCredential(strings.Split(azURL.Host, ".")[0], accessKey)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create Azure credential")
	}

	return httpURL, credential, nil
}

func getReinterpretFields(rootPath string, schemaRoot *schemaNode, noInterimLayer bool) map[string]ReinterpretField {
	reinterpretFields := make(map[string]ReinterpretField)
	for _, child := range schemaRoot.Children {
		currentPath := rootPath + "." + child.Name
		if child.Type == nil && child.ConvertedType == nil && child.NumChildren != nil {
			// STRUCT
			for k, v := range getReinterpretFields(currentPath, child, noInterimLayer) {
				reinterpretFields[k] = v
			}
			continue
		}

		if child.Type != nil && *child.Type == parquet.Type_INT96 {
			reinterpretFields[currentPath] = ReinterpretField{
				parquetType:   parquet.Type_INT96,
				convertedType: parquet.ConvertedType_TIMESTAMP_MICROS,
				precision:     0,
				scale:         0,
			}
			continue
		}

		if child.ConvertedType != nil {
			switch *child.ConvertedType {
			case parquet.ConvertedType_MAP, parquet.ConvertedType_LIST:
				if noInterimLayer {
					child = child.Children[0]
				}
				fallthrough
			case parquet.ConvertedType_MAP_KEY_VALUE:
				for k, v := range getReinterpretFields(currentPath, child, noInterimLayer) {
					reinterpretFields[k] = v
				}
			case parquet.ConvertedType_DECIMAL, parquet.ConvertedType_INTERVAL:
				reinterpretFields[currentPath] = ReinterpretField{
					parquetType:   *child.Type,
					convertedType: *child.ConvertedType,
					precision:     int(*child.Precision),
					scale:         int(*child.Scale),
				}
			}
		}
	}

	return reinterpretFields
}

func decimalToFloat(fieldAttr ReinterpretField, iface interface{}) (*float64, error) {
	if iface == nil {
		return nil, nil
	}

	switch value := iface.(type) {
	case int64:
		f64 := float64(value) / math.Pow10(fieldAttr.scale)
		return &f64, nil
	case int32:
		f64 := float64(value) / math.Pow10(fieldAttr.scale)
		return &f64, nil
	case string:
		buf := stringToBytes(fieldAttr, value)
		f64, err := strconv.ParseFloat(types.DECIMAL_BYTE_ARRAY_ToString(buf, fieldAttr.precision, fieldAttr.scale), 64)
		if err != nil {
			return nil, err
		}
		return &f64, nil
	}
	return nil, fmt.Errorf("unknown type: %s", reflect.TypeOf(iface))
}

func stringToBytes(fieldAttr ReinterpretField, value string) []byte {
	// INTERVAL uses LittleEndian, DECIMAL uses BigEndian
	// make sure all decimal-like value are all BigEndian
	buf := []byte(value)
	if fieldAttr.convertedType == parquet.ConvertedType_INTERVAL {
		for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
			buf[i], buf[j] = buf[j], buf[i]
		}
	}
	return buf
}

func newSchemaTree(reader *reader.ParquetReader) *schemaNode {
	schemas := reader.SchemaHandler.SchemaElements
	stack := []*schemaNode{}
	root := &schemaNode{
		SchemaElement: *schemas[0],
		Children:      []*schemaNode{},
	}
	stack = append(stack, root)

	pos := 1
	for len(stack) > 0 {
		node := stack[len(stack)-1]
		if len(node.Children) < int(node.SchemaElement.GetNumChildren()) {
			childNode := &schemaNode{
				SchemaElement: *schemas[pos],
				Children:      []*schemaNode{},
			}
			node.Children = append(node.Children, childNode)
			stack = append(stack, childNode)
			pos++
		} else {
			stack = stack[:len(stack)-1]
			if len(node.Children) == 0 {
				node.Children = nil
			}
		}
	}

	return root
}
