/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

// Package util provides ancillary functionality to go-vcloud-director library
// logging.go regulates logging for the whole library.
// See LOGGING.md for detailed usage
package util

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"runtime"
	"strings"
)

const (
	// Name of the environment variable that enables logging
	envUseLog = "GOVCD_LOG"

	// envOverwriteLog allows to overwrite file on every initialization
	envOverwriteLog = "GOVCD_LOG_OVERWRITE"

	// Name of the environment variable with the log file name
	envLogFileName = "GOVCD_LOG_FILE"

	// Name of the environment variable with the screen output
	envLogOnScreen = "GOVCD_LOG_ON_SCREEN"

	// Name of the environment variable that enables logging of passwords
	envLogPasswords = "GOVCD_LOG_PASSWORDS"

	// Name of the environment variable that enables logging of HTTP requests
	envLogSkipHttpReq = "GOVCD_LOG_SKIP_HTTP_REQ"

	// Name of the environment variable that enables logging of HTTP responses
	envLogSkipHttpResp = "GOVCD_LOG_SKIP_HTTP_RESP"

	// Name of the environment variable with a custom list of of responses to skip from logging
	envLogSkipTagList = "GOVCD_LOG_SKIP_TAGS"

	// Name of the environment variable with a custom list of of functions to include in the logging
	envApiLogFunctions = "GOVCD_LOG_FUNCTIONS"
)

var (
	// All go-vcloud director logging goes through this logger
	Logger *log.Logger

	// It's true if we're using an user provided logger
	customLogging bool = false

	// Name of the log file
	// activated by GOVCD_LOG_FILE
	ApiLogFileName string = "go-vcloud-director.log"

	// Globally enabling logs
	// activated by GOVCD_LOG
	EnableLogging bool = false

	// OverwriteLog specifies if log file should be overwritten on every run
	OverwriteLog bool = false

	// Enable logging of passwords
	// activated by GOVCD_LOG_PASSWORDS
	LogPasswords bool = false

	// Enable logging of Http requests
	// disabled by GOVCD_LOG_SKIP_HTTP_REQ
	LogHttpRequest bool = true

	// Enable logging of Http responses
	// disabled by GOVCD_LOG_SKIP_HTTP_RESP
	LogHttpResponse bool = true

	// List of tags to be excluded from logging
	skipTags = []string{"SupportedVersions", "ovf:License"}

	// List of functions included in logging
	// If this variable is filled, only operations from matching function names will be logged
	apiLogFunctions []string

	// Sends log to screen. If value is either "stderr" or "err"
	// logging will go to os.Stderr. For any other value it will
	// go to os.Stdout
	LogOnScreen string = ""

	// Flag indicating that a log file is open
	// logOpen bool = false

	// Text lines used for logging of http requests and responses
	lineLength int    = 80
	dashLine   string = strings.Repeat("-", lineLength)
	hashLine   string = strings.Repeat("#", lineLength)
)

func newLogger(logpath string) *log.Logger {
	var err error
	var file *os.File
	if OverwriteLog {
		file, err = os.OpenFile(logpath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0640)
	} else {
		file, err = os.OpenFile(logpath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0640)
	}

	if err != nil {
		fmt.Printf("error opening log file %s : %v", logpath, err)
		os.Exit(1)
	}
	return log.New(file, "", log.Ldate|log.Ltime)
}

func SetCustomLogger(customLogger *log.Logger) {
	Logger = customLogger
	EnableLogging = true
	customLogging = true
}

// initializes logging with known parameters
func SetLog() {
	if customLogging {
		return
	}
	if !EnableLogging {
		Logger = log.New(ioutil.Discard, "", log.Ldate|log.Ltime)
		return
	}

	// If no file name was set, logging goes to the screen
	if ApiLogFileName == "" {
		if LogOnScreen == "stderr" || LogOnScreen == "err" {
			log.SetOutput(os.Stderr)
			Logger = log.New(os.Stderr, "", log.Ldate|log.Ltime)
		} else {
			Logger = log.New(os.Stdout, "", log.Ldate|log.Ltime)
		}
	} else {
		Logger = newLogger(ApiLogFileName)
	}
	if len(skipTags) > 0 {
		Logger.Printf("### WILL SKIP THE FOLLOWING TAGS: %+v", skipTags)
	}
	if len(apiLogFunctions) > 0 {
		Logger.Printf("### WILL ONLY INCLUDE API LOGS FROM THE FOLLOWING FUNCTIONS: %+v", apiLogFunctions)
	}
}

// hidePasswords hides passwords that may be used in a request
func hidePasswords(in string, onScreen bool) string {
	if !onScreen && LogPasswords {
		return in
	}
	var out string
	re1 := regexp.MustCompile(`("[^\"]*[Pp]assword"\s*:\s*)"[^\"]+"`)
	out = re1.ReplaceAllString(in, `${1}"********"`)

	// Replace password in ADFS SAML request
	re2 := regexp.MustCompile(`(\s*<o:Password.*ext">)(.*)(</o:Password>)`)
	out = re2.ReplaceAllString(out, `${1}******${3}`)
	return out
}

// hideTokens hides SAML auth response token
func hideTokens(in string, onScreen bool) string {
	if !onScreen && LogPasswords {
		return in
	}
	var out string
	// Filters out the below:
	// Token data between <e:CipherValue> </e:CipherValue>
	re1 := regexp.MustCompile(`(.*<e:CipherValue>)(.*)(</e:CipherValue>.*)`)
	out = re1.ReplaceAllString(in, `${1}******${3}`)
	// Token data between <xenc:CipherValue> </xenc:CipherValue>
	re2 := regexp.MustCompile(`(.*<xenc:CipherValue>)(.*)(</xenc:CipherValue>.*)`)
	out = re2.ReplaceAllString(out, `${1}******${3}`)

	return out
}

// Determines whether a string is likely to contain binary data
func isBinary(data string, req *http.Request) bool {
	reContentRange := regexp.MustCompile(`(?i)content-range`)
	reMultipart := regexp.MustCompile(`(?i)multipart/form`)
	reMediaXml := regexp.MustCompile(`(?i)media+xml;`)
	for key, value := range req.Header {
		if reContentRange.MatchString(key) {
			return true
		}
		if reMultipart.MatchString(key) {
			return true
		}
		for _, v := range value {
			if reMediaXml.MatchString(v) {
				return true
			}
		}
	}
	return false
}

// SanitizedHeader returns a http.Header with sensitive fields masked
func SanitizedHeader(inputHeader http.Header) http.Header {
	if LogPasswords {
		return inputHeader
	}
	var sensitiveKeys = []string{
		"Config-Secret",
		"Authorization",
		"X-Vcloud-Authorization",
		"X-Vmware-Vcloud-Access-Token",
	}
	var sanitizedHeader = make(http.Header)
	for key, value := range inputHeader {
		// Explicitly mask only token in SIGN token so that other details are not obfuscated
		// Header format: SIGN token="`+base64GzippedSignToken+`",org="`+org+`"
		if (key == "authorization" || key == "Authorization") && len(value) == 1 &&
			strings.HasPrefix(value[0], "SIGN") && !LogPasswords {

			re := regexp.MustCompile(`(SIGN token=")([^"]*)(.*)`)
			out := re.ReplaceAllString(value[0], `${1}********${3}"`)

			Logger.Printf("\t%s: %s\n", key, out)
			// Do not perform any post processing on this header
			continue
		}

		for _, sk := range sensitiveKeys {
			if strings.EqualFold(sk, key) {
				value = []string{"********"}
				break
			}
		}
		sanitizedHeader[key] = value
	}
	return sanitizedHeader
}

// logSanitizedHeader logs the contents of the header after sanitizing
func logSanitizedHeader(inputHeader http.Header) {
	for key, value := range SanitizedHeader(inputHeader) {
		Logger.Printf("\t%s: %s\n", key, value)
	}
}

// Returns true if the caller function matches any of the functions in the include function list
func includeFunction(caller string) bool {
	if len(apiLogFunctions) > 0 {
		for _, f := range apiLogFunctions {
			reFunc := regexp.MustCompile(f)
			if reFunc.MatchString(caller) {
				return true
			}
		}
	} else {
		// If there is no include list, we include everything
		return true
	}
	// If we reach this point, none of the functions in the list matches the caller name
	return false
}

// Logs the essentials of a HTTP request
func ProcessRequestOutput(caller, operation, url, payload string, req *http.Request) {
	if !LogHttpRequest {
		return
	}
	if !includeFunction(caller) {
		return
	}

	Logger.Printf("%s\n", dashLine)
	Logger.Printf("Request caller: %s\n", caller)
	Logger.Printf("%s %s\n", operation, url)
	Logger.Printf("%s\n", dashLine)
	dataSize := len(payload)
	if isBinary(payload, req) {
		payload = "[binary data]"
	}
	if dataSize > 0 {
		Logger.Printf("Request data: [%d] %s\n", dataSize, hidePasswords(payload, false))
	}
	Logger.Printf("Req header:\n")
	logSanitizedHeader(req.Header)
}

// Logs the essentials of a HTTP response
func ProcessResponseOutput(caller string, resp *http.Response, result string) {
	if !LogHttpResponse {
		return
	}

	if !includeFunction(caller) {
		return
	}

	outText := result
	if len(skipTags) > 0 {
		for _, longTag := range skipTags {
			initialTag := `<` + longTag + `.*>`
			finalTag := `</` + longTag + `>`
			reInitialSearchTag := regexp.MustCompile(initialTag)

			// The `(?s)` flag treats the regular expression as a single line.
			// In this context, the dot matches every character until the next operator
			// The `.*?` is a non-greedy match of every character until the next operator, but
			// only matching the shortest possible portion.
			reSearchBothTags := regexp.MustCompile(`(?s)` + initialTag + `.*?` + finalTag)
			outRepl := fmt.Sprintf("[SKIPPING '%s' TAG AT USER'S REQUEST]", longTag)
			// We search for the initial long tag
			if reInitialSearchTag.MatchString(outText) {
				// If the first tag was found, we search the text to skip the whole output between the tags
				// Notice that if the second tag is not found, there won't be any replacement
				outText = reSearchBothTags.ReplaceAllString(outText, outRepl)
				break
			}
		}
	}
	Logger.Printf("%s\n", hashLine)
	Logger.Printf("Response caller %s\n", caller)
	Logger.Printf("Response status %s\n", resp.Status)
	Logger.Printf("%s\n", hashLine)
	Logger.Printf("Response header:\n")
	logSanitizedHeader(resp.Header)
	dataSize := len(result)
	outTextSize := len(outText)
	if outTextSize != dataSize {
		Logger.Printf("Response text: [%d -> %d] %s\n", dataSize, outTextSize, hideTokens(outText, false))
	} else {
		Logger.Printf("Response text: [%d] %s\n", dataSize, hideTokens(outText, false))
	}
}

// Sets the list of tahs to skip
func SetSkipTags(tags string) {
	if tags != "" {
		skipTags = strings.Split(tags, ",")
	}
}

// Sets the list of functions to include
func SetApiLogFunctions(functions string) {
	if functions != "" {
		apiLogFunctions = strings.Split(functions, ",")
	}
}

// Initializes default logging values
func InitLogging() {
	if os.Getenv(envLogSkipHttpReq) != "" {
		LogHttpRequest = false
	}

	if os.Getenv(envLogSkipHttpResp) != "" {
		LogHttpResponse = false
	}

	if os.Getenv(envApiLogFunctions) != "" {
		SetApiLogFunctions(os.Getenv(envApiLogFunctions))
	}

	if os.Getenv(envLogSkipTagList) != "" {
		SetSkipTags(os.Getenv(envLogSkipTagList))
	}

	if os.Getenv(envLogPasswords) != "" {
		EnableLogging = true
		LogPasswords = true
	}

	if os.Getenv(envLogFileName) != "" {
		EnableLogging = true
		ApiLogFileName = os.Getenv(envLogFileName)
	}

	LogOnScreen = os.Getenv(envLogOnScreen)
	if LogOnScreen != "" {
		ApiLogFileName = ""
		EnableLogging = true
	}

	if EnableLogging || os.Getenv(envUseLog) != "" {
		EnableLogging = true
	}

	if os.Getenv(envOverwriteLog) != "" {
		OverwriteLog = true
	}

	SetLog()
}

func init() {
	InitLogging()
}

// Returns the name of the function that called the
// current function.
// Used by functions that call processResponseOutput and
// processRequestOutput
func CallFuncName() string {
	fpcs := make([]uintptr, 1)
	n := runtime.Callers(3, fpcs)
	if n > 0 {
		fun := runtime.FuncForPC(fpcs[0] - 1)
		if fun != nil {
			return fun.Name()
		}
	}
	return ""
}

// Returns the name of the current function
func CurrentFuncName() string {
	fpcs := make([]uintptr, 1)
	runtime.Callers(2, fpcs)
	fun := runtime.FuncForPC(fpcs[0])
	return fun.Name()
}

// Returns a string containing up to 10 function names
// from the call stack
func FuncNameCallStack() string {
	// Gets the list of function names from the call stack
	fpcs := make([]uintptr, 10)
	runtime.Callers(0, fpcs)
	// Removes the function names from the reflect stack itself and the ones from the API management
	removeReflect := regexp.MustCompile(`^ runtime.call|reflect.Value|\bNewRequest\b|NewRequestWitNotEncodedParamsWithApiVersion|NewRequestWitNotEncodedParams|ExecuteRequest|ExecuteRequestWithoutResponse|ExecuteTaskRequest`)
	var stackStr []string
	// Gets up to 10 functions from the stack
	for N := 0; N < len(fpcs) && N < 10; N++ {
		fun := runtime.FuncForPC(fpcs[N])
		funcName := path.Base(fun.Name())
		if !removeReflect.MatchString(funcName) {
			stackStr = append(stackStr, funcName)
		}
	}
	// Reverses the function names stack, to make it easier to read
	var inverseStackStr []string
	for N := len(stackStr) - 1; N > 1; N-- {
		if stackStr[N] != "" && stackStr[N] != "." {
			inverseStackStr = append(inverseStackStr, stackStr[N])
		}
	}
	return strings.Join(inverseStackStr, "-->")
}
