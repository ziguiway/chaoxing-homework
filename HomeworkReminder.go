package main

import (
	"bytes"
	"fmt"
	"github.com/antchfx/htmlquery"
	"github.com/jordan-wright/email"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

var userAgentList = []string{"Mozilla/5.0 (compatible, MSIE 10.0, Windows NT, DigExt)",
	"Mozilla/4.0 (compatible, MSIE 7.0, Windows NT 5.1, 360SE)",
	"Mozilla/4.0 (compatible, MSIE 8.0, Windows NT 6.0, Trident/4.0)",
	"Mozilla/5.0 (compatible, MSIE 9.0, Windows NT 6.1, Trident/5.0,",
	"Opera/9.80 (Windows NT 6.1, U, en) Presto/2.8.131 Version/11.11",
	"Mozilla/4.0 (compatible, MSIE 7.0, Windows NT 5.1, TencentTraveler 4.0)",
	"Mozilla/5.0 (Windows, U, Windows NT 6.1, en-us) AppleWebKit/534.50 (KHTML, like Gecko) Version/5.1 Safari/534.50",
	"Mozilla/5.0 (Macintosh, Intel Mac OS X 10_7_0) AppleWebKit/535.11 (KHTML, like Gecko) Chrome/17.0.963.56 Safari/535.11",
	"Mozilla/5.0 (Macintosh, U, Intel Mac OS X 10_6_8, en-us) AppleWebKit/534.50 (KHTML, like Gecko) Version/5.1 Safari/534.50",
	"Mozilla/5.0 (Linux, U, Android 3.0, en-us, Xoom Build/HRI39) AppleWebKit/534.13 (KHTML, like Gecko) Version/4.0 Safari/534.13",
	"Mozilla/5.0 (iPad, U, CPU OS 4_3_3 like Mac OS X, en-us) AppleWebKit/533.17.9 (KHTML, like Gecko) Version/5.0.2 Mobile/8J2 Safari/6533.18.5",
	"Mozilla/4.0 (compatible, MSIE 7.0, Windows NT 5.1, Trident/4.0, SE 2.X MetaSr 1.0, SE 2.X MetaSr 1.0, .NET CLR 2.0.50727, SE 2.X MetaSr 1.0)",
	"Mozilla/5.0 (iPhone, U, CPU iPhone OS 4_3_3 like Mac OS X, en-us) AppleWebKit/533.17.9 (KHTML, like Gecko) Version/5.0.2 Mobile/8J2 Safari/6533.18.5",
	"MQQBrowser/26 Mozilla/5.0 (Linux, U, Android 2.3.7, zh-cn, MB200 Build/GRJ22, CyanogenMod-7) AppleWebKit/533.1 (KHTML, like Gecko) Version/4.0 Mobile Safari/533.1"}

const username = "15716216316"
const password = "123456"

type user struct {
	username string
	password string
	email    string
}

//课程类
type course struct {
	courseid    string
	clazzid     string
	personid    string
	id          string
	courseName  string
	teacherName string
}

type homework struct {
	courseName   string
	teacherName  string
	courseid     string
	clazzid      string
	homeworkTime string
	homeworkName string
	states       string
}

//消息类
type message struct {
	queryTime             string
	unfinishedAssignments []homework
}

func readUsernamePassword() []user {
	f, err := ioutil.ReadFile("./userInfo.txt")
	if err != nil {
		fmt.Println("read fail", err)
	}
	str := string(f)
	split := strings.Split(str, "#")
	var userList []user
	for _, each := range split {
		if each == "" {
			continue
		}
		var user user
		i := strings.Split(each, " ")
		user.username = i[0]
		user.password = i[1]
		user.email = i[2]
		userList = append(userList, user)
	}
	fmt.Println("", userList)
	return userList

}

//生成随机UserAgent
func GetRandomUserAgent() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return userAgentList[r.Intn(len(userAgentList))]
}

func getUrlRespHtml(url string, cookie string) string {

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("获取地址错误")
	}
	req.Header.Set("Cookie", cookie)
	req.Header.Add("Agent", GetRandomUserAgent())
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("登录错误")
	}
	resp_byte, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	respHtml := string(resp_byte)
	return respHtml

}

func getCookie(username string, password string) string {
	response, err := http.Get("https://passport2-api.chaoxing.com/v11/loginregister?code=" + password + "&cx_xxt_passport=json&uname=" + username + "&loginType=1&roleSelect=true")
	if err != nil {
		log.Println("该用户账号或密码输入错误")
	}
	defer response.Body.Close() //在回复后必须关闭回复的主体

	body, err := ioutil.ReadAll(response.Body)
	if err == nil {
		fmt.Println(string(body))
	}
	header := response.Header
	values := header.Values("Set-Cookie")

	test := ""
	for _, value := range values {
		test = test + value
		test = test + ";"
	}
	return test
}

func queryCourseInfo(html string) []course {
	parse, _ := htmlquery.Parse(strings.NewReader(html))
	//liList := htmlquery.Find(parse, "//li")
	liList := htmlquery.Find(parse, "/html/body/li")
	var course_list []course
	for _, li := range liList {
		var courseTemp = course{}

		courseName := htmlquery.FindOne(li, "//dt/text()").Data

		//teacherName := htmlquery.FindOne(li, "//dd/text()").Data

		attr := li.Attr
		courseTemp.courseid = attr[0].Val
		courseTemp.clazzid = attr[1].Val
		courseTemp.personid = attr[2].Val
		courseTemp.id = attr[3].Val
		courseTemp.courseName = courseName
		//courseTemp.teacherName = teacherName
		course_list = append(course_list, courseTemp)
	}
	return course_list
}

func queryHomeworkInfo(courseInfo []course, cookie string) []homework {
	var homework_list []homework

	for _, course := range courseInfo {

		url := "https://mooc1.chaoxing.com/work/task-list?courseId=" + course.courseid + "&classId=" + course.clazzid + "&vx=1"
		html := getUrlRespHtml(url, cookie)
		parse, _ := htmlquery.Parse(strings.NewReader(html))
		liList := htmlquery.Find(parse, "//li")
		var homeworkTemp = homework{}
		for _, li := range liList {
			homeworkTemp = homework{}
			var homeworkTime = ""
			homeworkName := htmlquery.FindOne(li, "//li//div/p/text()").Data
			find := htmlquery.Find(li, "//li//div/span/text()")

			state := find[0].Data

			if len(find) == 2 {
				homeworkTime = find[1].Data
			}

			homeworkTemp.homeworkTime = homeworkTime

			homeworkTemp.courseid = course.courseid
			homeworkTemp.clazzid = course.clazzid
			homeworkTemp.courseName = course.courseName
			homeworkTemp.teacherName = course.teacherName
			homeworkTemp.states = state
			homeworkTemp.homeworkName = homeworkName
			homework_list = append(homework_list, homeworkTemp)
			log.Println("正在查询" + homeworkTemp.courseName + ":")
			log.Println(homeworkTemp.courseid + "=" + homeworkTemp.clazzid + "=" + homeworkTemp.courseName + "=" + homeworkTemp.teacherName + "+" + homeworkTemp.states + "+" + homeworkTemp.homeworkName)
		}

	}

	return homework_list
}

func getUnfinishedAssignment(homeworkInfo []homework) message {
	var message message
	var unfinishedAssignments []homework
	for _, homework_each := range homeworkInfo {
		var UFHomework = homework{}

		if homework_each.states == "已完成" || homework_each.states == "待批阅" || homework_each.homeworkTime == "" {
			continue
		}
		UFHomework.homeworkTime = homework_each.homeworkTime
		UFHomework.teacherName = homework_each.teacherName
		UFHomework.courseName = homework_each.courseName
		UFHomework.homeworkName = homework_each.homeworkName
		unfinishedAssignments = append(unfinishedAssignments, UFHomework)

	}

	message.queryTime = time.Now().Format("2006-01-02 15:04:05")
	message.unfinishedAssignments = unfinishedAssignments
	return message
}

func sendMeg(message message, myemail string) bool {
	//message.unfinishedAssignments
	em := email.NewEmail()
	// 设置 sender 发送方 的邮箱 ， 此处可以填写自己的邮箱
	em.From = "ziguiway@163.com"

	// 设置 receiver 接收方 的邮箱  此处也可以填写自己的邮箱， 就是自己发邮件给自己
	em.To = []string{myemail}

	// 设置主题
	em.Subject = "学习通小助手提醒!你还有" + strconv.Itoa(len(message.unfinishedAssignments)) + "门作业未完成"

	var bt bytes.Buffer
	for index, assignment := range message.unfinishedAssignments {
		name := assignment.courseName
		homeworkName := assignment.homeworkName
		homeworkTime := assignment.homeworkTime
		bt.WriteString(strconv.Itoa(index+1) + "、")
		bt.WriteString(name + "\t")
		bt.WriteString(homeworkName + "\t")
		bt.WriteString(homeworkTime)
		bt.WriteString("\n\n")
	}

	em.Text = bt.Bytes()

	err := em.Send("smtp.163.com:25", smtp.PlainAuth("", "ziguiway@163.com", "CNQHUCPIYGYPPBPG", "smtp.163.com"))
	if err != nil {
		log.Fatal(err)
	}
	return true
}

func main() {
	//cookie := getCookie()

	users := readUsernamePassword()
	for index, u := range users {
		username := u.username
		password := u.password
		e := u.email
		log.Println("正在查询用户" + username + ",该用户位于用户" + strconv.Itoa(index+1))
		cookie := getCookie(username, password)
		respHtml := getUrlRespHtml("https://mooc1-2.chaoxing.com/course/phone/courselistdata?courseFolderId=0&isFiled=0&query=", cookie)
		courseInfo := queryCourseInfo(respHtml)
		homeworkInfo := queryHomeworkInfo(courseInfo, cookie)
		unfinishedAssignments := getUnfinishedAssignment(homeworkInfo)
		log.Println("", unfinishedAssignments)
		meg := sendMeg(unfinishedAssignments, e)
		if meg {
			log.Println("发送成功")
			time.Sleep(3000)
		}
	}
}
