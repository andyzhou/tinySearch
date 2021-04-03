package tc

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

/*
 * public util functions
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

//internal macro defines
const (
	UseLostConnectionErr = "use of closed network connection"
)

//time format
const (
	TimeLayOut = "2006-01-02 15:04:05" //can't be changed!!!
	Format              = "2006-01-02 15:04:05"
	GoFormat            = "2006-01-02 15:04:05.999999999"
	DateFormat          = "2006-01-02"
	FormattedDateFormat = "Jan 2, 2006"
	TimeFormat          = "15:04:05"
	HourMinuteFormat    = "15:04"
	HourFormat          = "15"
	DayDateTimeFormat   = "Mon, Aug 2, 2006 3:04 PM"
	CookieFormat        = "Monday, 02-Jan-2006 15:04:05 MST"
	RFC822Format        = "Mon, 02 Jan 06 15:04:05 -0700"
	RFC1036Format       = "Mon, 02 Jan 06 15:04:05 -0700"
	RFC2822Format       = "Mon, 02 Jan 2006 15:04:05 -0700"
	RFC3339Format       = "2006-01-02T15:04:05-07:00"
	RSSFormat           = "Mon, 02 Jan 2006 15:04:05 -0700"
)

//util info
type Utils struct {
}

//verify string is english, numeric or combination
func (u *Utils) VerifyEnglishNumeric(input string) bool {
	if input == "" {
		return false
	}
	for _, v := range input {
		if !unicode.IsLetter(v) && !unicode.IsNumber(v) {
			return false
		}
	}
	return true
}

//convert string slice to interface slice
func (u *Utils) ConvertStrSliceToGenSlice(orgSlice []string) []interface{} {
	if orgSlice == nil || len(orgSlice) <= 0 {
		return nil
	}
	result := make([]interface{}, 0)
	for _, v := range orgSlice {
		result = append(result, v)
	}
	return result
}


//reverse int slice
func (u *Utils) ReverseSlice(args ...interface{}) []interface{}{
	for i := 0; i < len(args)/2; i++ {
		j := len(args) - i - 1
		args[i], args[j] = args[j], args[i]
	}
	return args
}

//remove html tags
func (u *Utils) TrimHtml(src string, needLower bool) string {
	var (
		re *regexp.Regexp
	)

	if needLower {
		//convert to lower
		re, _ = regexp.Compile("\\<[\\S\\s]+?\\>")
		src = re.ReplaceAllStringFunc(src, strings.ToLower)
	}

	//remove style
	re, _ = regexp.Compile("\\<style[\\S\\s]+?\\</style\\>")
	src = re.ReplaceAllString(src, "")

	//remove script
	re, _ = regexp.Compile("\\<script[\\S\\s]+?\\</script\\>")
	src = re.ReplaceAllString(src, "")

	return strings.TrimSpace(src)
}

//shuffle slice
func (u *Utils) ShuffleSlice(data []int) {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(data), func(i, j int) { data[i], data[j] = data[j], data[i] })
}

//create json_array sql pass json data slice
func (u *Utils) GenJsonArrayAppendObject(
					tabField string,
					dataField string,
					jsonSlice []interface{},
				) (string, []interface{}) {
	var (
		buffer = bytes.NewBuffer(nil)
		tempStr string
		values = make([]interface{}, 0)
	)

	//basic check
	if tabField == "" || dataField == "" ||
	   jsonSlice == nil || len(jsonSlice) <= 0 {
		return buffer.String(), values
	}

	//check data field
	tempStr = fmt.Sprintf(", %s = JSON_SET(%s, '$.%s', IFNULL(%s->'$.%s',JSON_ARRAY()))",
						  tabField, tabField, dataField, tabField, dataField)
	buffer.WriteString(tempStr)

	//convert into relate data
	i := 0
	tempStr = fmt.Sprintf(", %s = JSON_ARRAY_APPEND(%s, ", tabField, tabField)
	buffer.WriteString(tempStr)
	for _, v := range jsonSlice {
		if i > 0 {
			buffer.WriteString(" ,")
		}
		tempStr = fmt.Sprintf("'$.%s', CAST(? AS JSON)", dataField)
		buffer.WriteString(tempStr)
		values = append(values, v)
		i++
	}
	buffer.WriteString(")")
	return buffer.String(), values
}

//create json_object sql pass json data map
//return subSql, values
func (u *Utils) GenJsonArray(
			valSlice []interface{},
		) (string, []interface{}) {
	var (
		buffer = bytes.NewBuffer(nil)
		values = make([]interface{}, 0)
	)
	//basic check
	if valSlice == nil || len(valSlice) <= 0 {
		return buffer.String(), values
	}

	arrayBuffer := bytes.NewBuffer(nil)
	arrayBuffer.WriteString("JSON_ARRAY(")
	for k, v2 := range valSlice {
		if k > 0 {
			arrayBuffer.WriteString(",")
		}
		arrayBuffer.WriteString("?")
		values = append(values, v2)
	}
	arrayBuffer.WriteString(")")
	return arrayBuffer.String(), values
}


//for general map
func (u *Utils) GenJsonObject(
					genHashMap map[string]interface{},
				) (string, []interface{}) {
	var (
		buffer = bytes.NewBuffer(nil)
		tempStr string
		values = make([]interface{}, 0)
	)

	//basic check
	if genHashMap == nil || len(genHashMap) <= 0 {
		return buffer.String(), values
	}

	//convert into relate data
	i := 0
	buffer.WriteString("json_object(")
	for k, v := range genHashMap {
		if i > 0 {
			buffer.WriteString(" ,")
		}
		//check value is array kind or not
		v1, isArray := v.([]string)
		if isArray {
			//format sub sql
			arrayBuffer := bytes.NewBuffer(nil)
			arrayBuffer.WriteString("JSON_ARRAY(")
			for k, v2 := range v1 {
				if k > 0 {
					arrayBuffer.WriteString(",")
				}
				arrayBuffer.WriteString("?")
				values = append(values, v2)
			}
			arrayBuffer.WriteString(")")

			//is array format
			tempStr = fmt.Sprintf("'%s', %s", k, arrayBuffer.String())
		}else{
			tempStr = fmt.Sprintf("'%s', ?", k)
			values = append(values, v)
		}
		buffer.WriteString(tempStr)
		i++
	}
	buffer.WriteString(")")

	return buffer.String(), values
}

//check or create dir
func (u *Utils) CheckOrCreateDir(dir string) error {
	_, err := os.Stat(dir)
	if err == nil {
		return err
	}
	bRet := os.IsExist(err)
	if bRet {
		return nil
	}
	err = os.Mkdir(dir, 0777)
	return err
}

//round float value
func (u *Utils) Round(f float64, n int) float64 {
	n10 := math.Pow10(n)
	return math.Trunc((f+0.5/n10)*n10) / n10
}

//reset object instance
func (u *Utils) RestObject(v interface{}) {
	p := reflect.ValueOf(v).Elem()
	p.Set(reflect.Zero(p.Type()))
}

//read byte file
//return bool and byte data
func (u *Utils) ReadBinFile(filePath string, needRemove bool) (bool, []byte) {
	if filePath == "" {
		return false, nil
	}

	//try read file
	byteData, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Println("ImageResize::ReadImage failed, err:", err.Error())
		return false, nil
	}

	if needRemove {
		os.Remove(filePath)
	}

	return true, byteData
}

//convert seconds to time string format
func (u *Utils) Seconds2TimeStr(seconds int) string {
	var (
		hourStr, minuteStr, secondStr string
	)

	if seconds <= 0 {
		return ""
	}

	hourInt := seconds / 3600
	minuteInt := (seconds - hourInt * 3600) / 60
	secondInt := seconds - hourInt * 3600 - minuteInt * 60

	if hourInt > 0 {
		if hourInt > 9 {
			hourStr = fmt.Sprintf("%d:", hourInt)
		}else{
			hourStr = fmt.Sprintf("0%d:", hourInt)
		}
	}

	if minuteInt > 9 {
		minuteStr = fmt.Sprintf("%d", minuteInt)
	}else{
		minuteStr = fmt.Sprintf("0%d", minuteInt)
	}

	if secondInt > 9 {
		secondStr = fmt.Sprintf("%d", secondInt)
	}else{
		secondStr = fmt.Sprintf("0%d", secondInt)
	}

	//format time string
	timeStr := fmt.Sprintf("%s%s:%s", hourStr, minuteStr, secondStr)
	return timeStr
}

//convert time string format to int format
func (u *Utils) TimeStr2Seconds(timeStr string) int {
	var (
		timeSeconds int
		tempIntVal int
		tempFloatVal float64
	)

	//basic check
	if timeStr == "" {
		return timeSeconds
	}

	i := 1
	tempSlice := strings.Split(timeStr, ":")
	for _, info := range tempSlice {
		switch i {
		case 1://hour
			tempIntVal, _ = strconv.Atoi(info)
			timeSeconds += tempIntVal * 3600
		case 2://minute
			tempIntVal, _ = strconv.Atoi(info)
			timeSeconds += tempIntVal * 60
		case 3:	//second
			tempFloatVal, _ = strconv.ParseFloat(info, 64)
			timeSeconds += int(tempFloatVal)
		}
		i++
	}

	return timeSeconds
}


//upper first character
func (u *Utils) UcFirst(str string) string {
	for i, v := range str {
		return string(unicode.ToUpper(v)) + str[i+1:]
	}
	return ""
}

//lower first character
func (u *Utils) LcFirst(str string) string {
	for i, v := range str {
		return string(unicode.ToLower(v)) + str[i+1:]
	}
	return ""
}

//sub string, support utf8 string
func (u *Utils) SubString(source string, start int, length int) string {
	rs := []rune(source)
	len := len(rs)
	if start < 0 {
		start = 0
	}
	if start >= len {
		start = len
	}
	end := start + length
	if end > len {
		end = len
	}
	return string(rs[start:end])
}

//generate md5 string
func (u *Utils) GenMd5(orgString string) string {
	if len(orgString) <= 0 {
		return ""
	}
	m := md5.New()
	m.Write([]byte(orgString))
	return hex.EncodeToString(m.Sum(nil))
}

//get rand number
func (u *Utils) GetRandomVal(maxVal int) int {
	randSand := rand.NewSource(time.Now().UnixNano())
	r := rand.New(randSand)
	return r.Intn(maxVal)
}

func (u *Utils) GetRealRandomVal(maxVal int) int {
	return int(rand.Float64() * 1000) % (maxVal)
}

//get current date, like YYYY-MM-DD
func (u *Utils) GetCurDate() string {
	now := time.Now()
	curDate := fmt.Sprintf("%d-%d-%d", now.Year(), now.Month(), now.Day())
	return curDate
}

//get current month, like YYYYMM
func (u *Utils) GetCurMonthInt() int {
	var (
		monthStr string
	)
	now := time.Now()
	year := now.Year()
	month := now.Month()
	if month > 9 {
		monthStr = fmt.Sprintf("%d", month)
	}else{
		monthStr = fmt.Sprintf("0%d", month)
	}
	finalDate := fmt.Sprintf("%d%s", year, monthStr)
	finalDateInt, _ := strconv.Atoi(finalDate)
	return finalDateInt
}

//get current day, like YYYYMMDD
func (u *Utils) GetCurDateInt() int {
	var (
		dayStr string
	)
	now := time.Now()
	day := now.Day()
	if day > 9 {
		dayStr = fmt.Sprintf("%d", day)
	}else{
		dayStr = fmt.Sprintf("0%d", day)
	}
	finalDate := fmt.Sprintf("%d%s", u.GetCurMonthInt(), dayStr)
	finalDateInt, _ := strconv.Atoi(finalDate)
	return finalDateInt
}

//get current hour, like YYYYMMDDHH
func (u *Utils) GetCurHourInt() int {
	var (
		hourStr string
	)
	now := time.Now()
	hour := now.Hour()
	if hour > 9 {
		hourStr = fmt.Sprintf("%d", hour)
	}else{
		hourStr = fmt.Sprintf("0%d", hour)
	}
	finalHour := fmt.Sprintf("%d%s", u.GetCurDateInt(), hourStr)
	finalHourInt, _ := strconv.Atoi(finalHour)
	return finalHourInt
}

//get current minute, like YYYYMMDDHHMM
func (u *Utils) GetCurMinuteInt() int {
	var (
		minuteStr string
	)
	now := time.Now()
	minute := now.Minute()
	if minute > 9 {
		minuteStr = fmt.Sprintf("%d", minute)
	}else{
		minuteStr = fmt.Sprintf("0%d", minute)
	}
	finalMinute := fmt.Sprintf("%d%s", u.GetCurDateInt(), minuteStr)
	finalMinuteInt, _ := strconv.Atoi(finalMinute)
	return finalMinuteInt
}

//convert date time string to timestamp
func (u *Utils) DateTime2Unix(dateTime string) int64 {
	//remove un useful info
	dateTime = strings.Replace(dateTime, "T", " ", -1)
	dateTime = strings.Replace(dateTime, "Z", "", -1)

	//theTime, err := time.Parse(TimeLayOut, dateTime)
	theTime, err := time.ParseInLocation(TimeLayOut, dateTime, time.Local)
	if err != nil {
		log.Println("TimeToUnixStamp, convert failed, err:", err.Error())
		return 0
	}
	return theTime.Unix()
}

//convert timestamp to date format, like YYYY-MM-DD
func (u *Utils) TimeStamp2Date(timeStamp int64) string {
	dateTime := time.Unix(timeStamp, 0).Format(TimeLayOut)
	tempSlice := strings.Split(dateTime, " ")
	if tempSlice == nil || len(tempSlice) <= 0 {
		return ""
	}
	return tempSlice[0]
}

//convert timestamp to data time string format
func (u *Utils) TimeStamp2DateTime(timeStamp int64) string {
	return time.Unix(timeStamp, 0).Format(TimeLayOut)
}

//convert timestamp like 'Oct 10, 2020' format
func (u *Utils) TimeStampToDayStr(timeStamp int64) string {
	date := u.TimeStamp2Date(timeStamp)
	if date == "" {
		return  ""
	}
	tempSlice := strings.Split(date, "-")
	if tempSlice == nil || len(tempSlice) < 3 {
		return ""
	}
	year := tempSlice[0]
	month, _ := strconv.Atoi(tempSlice[1])
	day := tempSlice[2]
	return fmt.Sprintf("%s %s, %s", time.Month(month).String(), day, year)
}

//check tcp error
//return true need quit, false just gen error
func (u *Utils) CheckTcpError(err error) bool {
	var (
		isOk bool
		netOpError *net.OpError
	)
	if err != nil {
		log.Println("service::checkTcpError, err:", err)
		netOpError, isOk = err.(*net.OpError)
		if isOk && netOpError.Err.Error() == UseLostConnectionErr {
			//use a broken connect
			return true
		}
		if err == io.EOF {
			//ignore EOF since client might send nothing for the moment
			return true
		}
		netErr, ok := err.(net.Error)
		if ok && netErr.Timeout() {
			//socket operate time out
			return true
		}
	}
	return false
}

//get current host
func (u *Utils) GetCurHost() string {
	//get local ip
	var defaultHost = "127.0.0.1"
	var ipAddress string

	addressSlice, err := net.InterfaceAddrs()
	if nil != err {
		log.Fatal("Get local IP addr failed!!!")
		return defaultHost
	}
	for _, address := range addressSlice {
		if ipNet, ok := address.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if nil != ipNet.IP.To4() {
				ipAddress = ipNet.IP.String()
				return ipAddress
			}
		}
	}
	return defaultHost
}

//add parameter into slice
func (u *Utils) AddParamInSlice(orgSlice *[]string, para string) bool {
	if para == "" {
		return false
	}
	*orgSlice = append(*orgSlice, para)
	return true
}

//convert slice to string
func (u *Utils) Slice2Str(orgSlice []string) string {
	var result string
	if len(orgSlice) <= 0 {
		return result
	}
	for _, v := range orgSlice {
		result += v
	}
	return result
}

//check file stat and last modify time
func (u *Utils) GetFileModifyTime(filePath string) int64 {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0
	}
	modifyTime := fileInfo.ModTime().Unix()
	return modifyTime
}

//get file info
func (u *Utils) GetFileInfo(filePath string) os.FileInfo {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil
	}
	return fileInfo
}

//gather all ip from client
func (u *Utils) GetClientAllIp(r *http.Request) []string {
	var tempStr string
	var ipSlice = make([]string, 0)

	//get original data
	clientAddress := r.RemoteAddr
	xRealIp := r.Header.Get("X-Real-IP")
	xForwardedFor := r.Header.Get("X-Forwarded-For")

	//analyze general ip
	if clientAddress != "" {
		tempStr = u.analyzeClientIp(clientAddress)
		if tempStr != "" {
			ipSlice = append(ipSlice, tempStr)
		}
	}

	//analyze x-real-ip
	if xRealIp != "" {
		tempStr = u.analyzeClientIp(clientAddress)
		if tempStr != "" {
			ipSlice = append(ipSlice, tempStr)
		}
	}

	//analyze x-forward-for
	//like:192.168.0.1,192.168.0.2
	if xForwardedFor != "" {
		tempSlice := strings.Split(xForwardedFor, ",")
		if len(tempSlice) > 0 {
			for _, tmpAddr := range tempSlice {
				tempStr = u.analyzeClientIp(tmpAddr)
				if tempStr != "" {
					ipSlice = append(ipSlice, tempStr)
				}
			}
		}
	}

	return ipSlice
}

//analyze client ip
func (u *Utils) analyzeClientIp(address string) string {
	tempSlice := strings.Split(address, ":")
	if len(tempSlice) < 1 {
		return ""
	}
	return tempSlice[0]
}
