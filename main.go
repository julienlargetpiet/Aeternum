package main

import (
  "fmt"
  "net/http"
  "html/template"
  "database/sql"
  "os"
  "time"
  "crypto/aes"
  _"github.com/go-sql-driver/mysql"
)


var bgn_year int32 = 2024
var max_year int32 = 2124 //100 years delta max

var templates = template.Must(template.ParseFiles("templates/track_page.html", "templates/choose.html", "templates/see_day.html", "templates/provide_page.html", "templates/calculate.html", "templates/result.html"))
const secret_key string = "N1PCdw3M2B1TfJhoaY2mL736p2vCUc47"
var month_days = [12]uint32{31, 28, 31, 30, 31, 30, 
                            31, 31, 30, 31, 30, 31}
var month_days_leap = [12]uint32{31, 29, 31, 30, 31, 30, 
                                 31, 31, 30, 31, 30, 31}
var str_month_days = [12]string{"31", "28", "31", "30", "31", "30", 
                               "31", "31", "30", "31", "30", "31"}
var str_month_days_leap = [12]string{"31", "29", "31", "30", "31", 
                                "30", "31", "31", "30", "31", "30", "31"}
var str_month_nb = [12]string{"01", "02", "03", "04", "05", "06", 
                             "07", "08", "09", "10", "11", "12"}
var str_day_nb = [31]string{"1", "2", "3", "4", "5", "6", "7", 
                            "8", "9", "10", "11", "12",
                            "13", "14", "15", "16", "17", "18", "19", 
                            "20", "21", "22", "23", "24",
                            "25", "26", "27", "28", "29", "30", "31"}
var str_hour_nb = [24]string{"1", "2", "3", "4", "5", "6", "7", 
                             "8", "9", "10", "11", "12", 
                             "13", "14", "15", "16", "17", "18", "19", 
                             "20", "21", "22", "23", "24"}
var str_minute_nb = [60]string{"00", "01", "02", "03", "04", "05", "06", 
                            "07", "08", "09", "10", "11", "12",
                            "13", "14", "15", "16", "17", "18", "19", 
                            "20", "21", "22", "23", "24",
                            "25", "26", "27", "28", "29", "30", "31",
                            "32", "33", "34", "35", "36", "37", "38",
                            "39", "40", "41", "42", "43", "44", "45",
                            "46", "47", "48", "49", "50", "51", "52",
                            "53", "54", "55", "56", "57", "58", "59"}
var ref_nb = []uint8{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}
var ref_ltr = [52]uint8{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z'}
var ref_spechr = [24]uint8{'!', '.', ':', ';', '\\', '-', '%', '*', ',', '_', '/', '<', '>', '=', '[', ']', '\'', '{', '}', '[', ']', '(', ')', '"'}
var ref_temp_password = [11]uint8{'-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}
var banned_char_username = [22]uint8{'_', ' ', '/', '?', '$', 
                           '&', '@', '#', '.', ',', '\\', '|', 
                           '{', '}', '(', ')', '^', '<', '>', '%', ':'}
var only_usrs bool = false
var banned_usernames = [10]string{"Root", "ROOT", "root", "Admin", "ADMIN", "admin", "main", "static", "templates", "main.go"}

func Int32ToString(x *int32) string {
  const base int32 = 10
  var remainder int32
  rtn_str := ""
  for *x > 0 {
    remainder = *x % base
    rtn_str = string(remainder + 48) + rtn_str
    *x -= remainder
    *x /= 10
  }
  return rtn_str
}

func Int64ToString(x *int64) string {
  const base int64 = 10
  var remainder int64
  rtn_str := ""
  for *x > 0 {
    remainder = *x % base
    rtn_str = string(remainder + 48) + rtn_str
    *x -= remainder
    *x /= 10
  }
  return rtn_str
}

func StringToInt32(x string) int32 {
  var rtn_val int32 = 0
  var lngth int = len(x)
  var i2 int32
  var cur_rn uint8
  var i int
  for i = 0; i + 1 < lngth; i++ {
    cur_rn = x[i]
    i2 = 0
    for i2 < 11 && cur_rn != ref_nb[i2]{
      i2++
    }
    rtn_val += i2
    rtn_val *= 10
  }
  cur_rn = x[i]
  i2 = 0
  for cur_rn != ref_nb[i2]{
    i2++
  }
  rtn_val += i2
  return rtn_val
}

func URLToCredentials(url string) (string, string, bool) {
  rtn_str := ""
  rtn_str2 := ""
  var i int = len(url) - 1
  var i3 int
  n_ref := i
  var cur_bool bool
  for url[i] != '_' {
    cur_bool = false
    for i3 = 0; i3 < 11; i3++ {
      if url[i] == ref_temp_password[i3] {
        cur_bool = true
        break
      }
    }
    if !cur_bool {
      return "", "", false
    }
    rtn_str += string(url[i])
    i--
  }
  if i == n_ref {
    return "", "", false
  }
  i--
  for url[i] != '_' {
    for i3 = 0; i3 < 3; i3++ {
      if url[i] == banned_char_username[i3] {
        return "", "", false
      }
    }
    rtn_str2 += string(url[i])
    i--
  }
  i = 0
  var n int = len(rtn_str)
  password_rune := []rune(rtn_str)
  var tmp_val rune
  for i < n / 2 {
    tmp_val = password_rune[i]
    password_rune[i] = password_rune[n - 1 - i]
    password_rune[n - 1 - i] = tmp_val
    i++
  }
  username_rune := []rune(rtn_str2)
  n = len(rtn_str2)
  i = 0
  for i < n / 2 {
    tmp_val = username_rune[i]
    username_rune[i] = username_rune[n - 1 - i]
    username_rune[n - 1 - i] = tmp_val
    i++
  }
  return string(password_rune), string(username_rune), true
}

func CredentialsToURL(tmp_password string, username *string, db *sql.DB) (string, error) {
  cur_time := time.Now().Unix()
  string_time := Int64ToString(&cur_time)
  string_time = string_time[len(string_time) - 4:]
  tmp_password = tmp_password[:12]
  tmp_password += string_time
  
  aes, err := aes.NewCipher([]byte(secret_key))
  if err != nil {
    return "", err
  }
  
  ciphered_password := make([]byte, 16)
  aes.Encrypt(ciphered_password, []byte(tmp_password))
  
  p_rotated_link := ""
  var cur_str string
  var cur_rune int32
  var rotated_link string
  
  for i := 0; i < 16; i++ {
    cur_rune = int32(ciphered_password[i])
    cur_str = Int32ToString(&cur_rune) + "-"
    p_rotated_link += cur_str
  }
  
  _, err = db.Exec("UPDATE credentials SET temp_password=? WHERE username=?;", p_rotated_link, *username)
  rotated_link = "_" + *username + "_" + p_rotated_link

  return rotated_link, nil
}

func GoodUsername(given_username string) bool {
  if len(given_username) == 0 {
    return false
  }
  var i2 int
  var cur_val uint8
  for i:= 0; i < len(given_username); i++ {
    cur_val = given_username[i]
    for i2 = 0; i2 < 22; i2++ {
      if cur_val == banned_char_username[i2] {
        return false
      }
    }
  }
  for _, usr := range banned_usernames {
    if given_username == usr {
      return false
    }
  }
  var is_in = false
  if only_usrs {
    for _, usr := range banned_usernames {
      if given_username == usr {
        is_in = true
        break
      }
    }
    if !is_in {
      return false
    }
  }
  return true 
}

func GoodPassword(given_password string) bool {
  var n int = len(given_password)
  if n != 16 {
    return false
  }
  var i uint = 0
  var i2 uint
  var cur_val uint8
  var agn bool = true
  for agn && i < 16 {
    cur_val = given_password[i]
    i2 = 0
    for i2 < 10 && cur_val != ref_nb[i2] {
      i2++
    }
    if i2 < 10 {
      agn = false
    }
    i++
  }
  if agn {
    return false
  }
  agn = true
  i = 0
  for agn && i < 16 {
    cur_val = given_password[i]
    i2 = 0
    for i2 < 52 && cur_val != ref_ltr[i2] {
      i2++
    }
    if i2 < 52 {
      agn = false
    }
    i++
  }
  i = 0
  if agn {
    return false
  }
  agn = true
  for agn && i < 16 {
    cur_val = given_password[i]
    i2 = 0
    for i2 < 24 && cur_val != ref_spechr[i2] {
      i2++
    }
    if i2 < 24 {
      agn = false
    }
    i++
  }
  if agn {
    return false
  }
  return true
}

func EvaluateConnectionPassword(given_password *string, username *string, db *sql.DB) bool {
  var real_password string
  username_query := db.QueryRow("SELECT password FROM credentials WHERE BINARY username = ?;", username)
  err := username_query.Scan(&real_password)
  if err != nil {
    fmt.Println(err)
    return false
  }
  if real_password != *given_password {
    return false
  }
  return true
}

func EvaluatePassword(given_password *string, username *string, db *sql.DB) bool {
  var real_password string
  username_query := db.QueryRow("SELECT temp_password FROM credentials WHERE username = ?;", username)
  err := username_query.Scan(&real_password)
  if err != nil {
    return false
  }
  if real_password != *given_password {
    return false
  }
  return true
}

func is_leap(x *int32) bool {
  if *x % 4 == 0 {
    if *x % 100 == 0 {
      if *x % 400 == 0 {
        return true
      } else {
        return false
      }
    } else {
      return true
    }
  } else {
    return false
  }
}

func DateToScds(x string, rotated_link *string) (uint32, bool, string) {
  if len(x) == 0 {
    return 0, false, ""
  }
  var rtn_scds uint32
  rtn_scds = 0
  year := ""
  month := ""
  day := ""
  var i int32
  var n int32 = int32(len(x))
  i = n - 1
  for i > -1 && x[i] != '-' {
    year = string(x[i]) + year
    i--
  }
  if i == -1 {
    return 0, false, ""
  }
  is_valid, resp := EvaluateYear(year, rotated_link)
  if !is_valid {
    return 0, false, resp
  }
  i--
  for i > -1 && x[i] != '-' {
    month = string(x[i]) + month
    i--
  }
  if i == -1 {
    return 0, false, ""
  }
  is_valid, resp = EvaluateMonth(&month, &year, rotated_link)
  if !is_valid {
    return 0, false, resp
  }
  i--
  for i > -1 {
    day = string(x[i]) + day
    i--
  }
  int_val := StringToInt32(year)
  i = bgn_year
  var leap_vl bool
  for i < int_val {
    leap_vl = is_leap(&i)
    if !leap_vl {
      rtn_scds += 365 * 24 * 3600
    } else {
      rtn_scds += 366 * 24 * 3600
    }
    i++
  }
  leap_vl = is_leap(&i)
  is_valid, resp = EvaluateDay(&day, &month, &leap_vl, rotated_link)
  if !is_valid {
    return 0, false, resp
  }
  i = 1
  int_val = StringToInt32(month)
  if !leap_vl {
    for i < int_val {
      rtn_scds += month_days[i] * 24 * 360
      i++
    }
  } else {
    for i < int_val {
      rtn_scds += month_days_leap[i] * 24 * 3600
      i++
    }
  }
  i = 1
  int_val = StringToInt32(day)
  for i <= int_val {
    rtn_scds += 24 * 3600
    i++
  }
  return rtn_scds, true, ""
}

func EvaluateDay(cur_day *string,
                   cur_month *string,
                   leap_vl *bool, 
                   rotated_link *string) (bool, string) {
  if len(*cur_day) == 0 {
    return false, `<b>This day does not exist</b><br><a href="../choose/` + *rotated_link + `">Choose</a>`
  }
  var i uint32 = 0
  var ref_i uint32
  for str_month_nb[i] != *cur_month {
    i++
  }
  ref_i = i
  i = 0 
  if !*leap_vl {
    for i < month_days[ref_i] {
      if *cur_day == str_day_nb[i] {
        break
      }
      i++
    }
  } else {
    for i < month_days_leap[ref_i] {
      if *cur_day == str_day_nb[i] {
        break
      }
      i++
    }
  }
  if i == 31 {
    return false, `<b>This day does not exist</b><br><a href="../choose/` + *rotated_link + `">Choose</a>`
  }
  return true, ""
}

func EvaluateMonth(cur_month *string,
                   cur_year *string, 
                   rotated_link *string) (bool, string) {
  for _, mnth := range str_month_nb {
    if mnth == *cur_month {
      return true, ""
    }
  }
  return false, `<b>This month does not exist</b><br><a href="../choose/` + *rotated_link + `">Choose</a>`
}

func EvaluateYear(cur_year string, rotated_link *string) (bool, string) {
  var rtn_str string
  if len(cur_year) == 0 {
    rtn_str = `<b>Year has no value</b><br><a href="../choose/` + *rotated_link + `">Go Back</a>`
    return false, rtn_str
  }
  if len(cur_year) > 4 { //extra check because year is encoded on int32 consider switching to StringToInt64() if the year is too large
    rtn_str = `<b>For admin: If we are far in the future (> 9999), change the condition at line 556 from 4 to 5</b><br><a href="../choose/` + *rotated_link + `">Go Back</a>`
    return false, rtn_str
  }
  Cur_Year := StringToInt32(cur_year)
  if Cur_Year > max_year {
    rtn_str = `<b>The year is greater than max_year</b><br><a href="../choose/` + *rotated_link + `">Go Back</a>`
    return false, rtn_str
  } else if Cur_Year < bgn_year {
    rtn_str = `<b>The year is lower than bgn_year</b><br><a href="../choose/` + *rotated_link + `">Go Back</a>`
    return false, rtn_str
  }
  return true, ""
}

func EvaluateWorkHours(data string, rotated_link *string) (float32, bool , string) {
  n := len(data)
  if n == 0{
    return 0, false, `<b>Input not valid</b><br><a href="../choose/` + *rotated_link + `">Choose</a>`
  }
  data += ";"
  n += 1
  var cur_hour string
  var cur_min string
  var i2 int
  var bgn_hour int32
  var bgn_min int32
  var bgn_hour_val float32
  var end_hour int32
  var end_min int32
  var rtn_val float32
  var end_hour_val float32
  rtn_val = 0
  var i int = 0
  for i < n {
    cur_hour = ""
    cur_min = ""
    for i < n && data[i] != 'h' {
      cur_hour += string(data[i])
      i++
    }
    if i == n {
      return 0, false, `<b>Input not valid</b><br><a href="../choose/` + *rotated_link + `">Choose</a>`
    }
    i2 = 0
    for i2 < 24 {
      if cur_hour == str_hour_nb[i2] {
        break
      }
      i2++
    }
    if i2 == 24 {
      return 0, false, `<b>Input not valid</b><br><a href="../choose/` + *rotated_link + `">Choose</a>`
    }
    i++
    for i < n && data[i] != '-' {
      cur_min += string(data[i])
      i++
    }
    if i == n {
      return 0, false, `<b>Input not valid</b><br><a href="../choose/` + *rotated_link + `">Choose</a>`
    }
    i2 = 0
    for i2 < 60 {
      if cur_min == str_minute_nb[i2] {
        break
      }
      i2++
    }
    if i2 == 60 {
      return 0, false, `<b>Input not valid</b><br><a href="../choose/` + *rotated_link + `">Choose</a>`
    }
    bgn_hour = StringToInt32(cur_hour)
    bgn_min = StringToInt32(cur_min)
    bgn_hour_val = float32(bgn_hour) + float32(bgn_min) / 60 
    i++
    cur_hour = ""
    cur_min = ""
    for i < n && data[i] != 'h' {
      cur_hour += string(data[i])
      i++
    }
    if i == n {
      return 0, false, `<b>Input not valid</b><br><a href="../choose/` + *rotated_link + `">Choose</a>`
    }
    i2 = 0
    for i2 < 24 {
      if cur_hour == str_hour_nb[i2] {
        break
      }
      i2++
    }
    if i2 == 24 {
      return 0, false, `<b>Input not valid</b><br><a href="../choose/` + *rotated_link + `">Choose</a>`
    }
    i++
    for i < n && data[i] != ';' {
      cur_min += string(data[i])
      i++
    }
    if i == n {
      return 0, false, `<b>Input not valid</b><br><a href="../choose/` + *rotated_link + `">Choose</a>`
    }
    i2 = 0
    for i2 < 60 {
      if cur_min == str_minute_nb[i2] {
        break
      }
      i2++
    }
    if i2 == 60 {
      return 0, false, `<b>Input not valid</b><br><a href="../choose/` + *rotated_link + `">Choose</a>`
    }
    end_hour = StringToInt32(cur_hour)
    end_min = StringToInt32(cur_min)
    end_hour_val = float32(end_hour) + float32(end_min) / 60
    i++
    if end_hour_val <= bgn_hour_val {
      return 0, false, `<b>Begin time can't be greater or equal than end time</b><br><a href="../choose/` + *rotated_link + `">Choose</a>`
    }
    rtn_val += (end_hour_val - bgn_hour_val)
  }
  return rtn_val, true, ""
}

func ConnectDatabase() (*sql.DB, error) {
  var credentials = "kvv:1234@(localhost:3306)/Aeternum"
  db, err := sql.Open("mysql", credentials)
  if err != nil {
    return nil, err
  }
  return db, nil
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
  if r.Method != "GET" {
    w.Write([]byte("<b>Bad Method</b>"))
    return
  }
  if r.URL.Path != "/" {
    w.Write([]byte("<b>Bad URL</b>"))
    return
  }
  data, err := os.ReadFile("templates/index.html")
  if err != nil {
    fmt.Println(err)
    w.Write([]byte("<b>Something went wrong</b>"))
    return
  }
  w.Write(data)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
  if r.Method != "GET" {
    w.Write([]byte("<b>Bad Method</b>"))
    return
  }
  if r.URL.Path != "/login/" {
    w.Write([]byte("<b>Bad URL</b>"))
    return
  }
  data, err := os.ReadFile("templates/login.html")
  if err != nil {
    fmt.Println(err)
    w.Write([]byte("<b>Something went wrong</b>"))
    return
  }
  w.Write(data)
}

func ConnectionHandler(db *sql.DB) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    if r.Method != "POST" {
      w.Write([]byte("<b>Bad Method</b>"))
      return
    }
    
    if r.URL.Path != "/connection/" {
      w.Write([]byte("<b>Bad URL</b>"))
      return
    }

    username_form := r.FormValue("username")
    password_form := r.FormValue("password")
    
    rtn_bool := EvaluateConnectionPassword(&password_form, 
                                           &username_form, 
                                           db)

    if !rtn_bool {
      w.Write([]byte("<b>Not allowed to be here</b>"))
      return
    } else {
      rotated_link, err := CredentialsToURL(password_form, 
                                        &username_form, 
                                        db) 
      if err != nil {
        fmt.Println(err)
        return
      }
     http.Redirect(w, r, "/choose/" + rotated_link, http.StatusFound)
    }
  }
}

func CreateAccountHandler(w http.ResponseWriter, r *http.Request) {
  if r.Method != "GET" {
    w.Write([]byte("<b>Bad Method</b>"))
    return
  }
  if r.URL.Path != "/create_account/" {
    w.Write([]byte("<b>Bad URL</b>"))
    return
  }
  data, err := os.ReadFile("templates/create_account.html")
  if err != nil {
    fmt.Println(err)
    w.Write([]byte("<b>Something went wrong</b>"))
    return
  }
  w.Write(data)
}

func NewAccountHandler(db *sql.DB) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    if r.Method != "POST" {
      w.Write([]byte("<b>Bad Method</b>"))
      return
    }
    if r.URL.Path != "/new_account/" {
      w.Write([]byte("<b>Bad URL</b>"))
      return
    }
    username := r.FormValue("username")
    is_valid := GoodUsername(username)
    if !is_valid {
      w.Write([]byte("<b>Invalid username</b>"))
      return
    }
    password := r.FormValue("password")
    is_valid = GoodPassword(password)
    if !is_valid {
      w.Write([]byte("<b>Invalid password</b>"))
      return
    }
    var alrd_username string
    content := db.QueryRow("SELECT username FROM credentials WHERE username=?;", username)
    err := content.Scan(&alrd_username)
    if err == nil {
      w.Write([]byte("<b>Username already taken</b>"))
      return
    }
    _, err = db.Exec("INSERT INTO credentials VALUE(?, ?, ' ');",
                              username, password)
    if err != nil {
      w.Write([]byte("<b>Something went wrong</b>"))
      return
    }
    _, err = db.Exec("CREATE TABLE " + username + " (name VARCHAR(12), nb_seconds BIGINT, work_hours DOUBLE);")
    if err != nil {
      w.Write([]byte("<b>Something went wrong</b>"))
      return
    }
    http.Redirect(w, r, "/login/", http.StatusFound)
  }
}

type ChooseStruct struct {
  NextURL string
}

func ChooseHandler(db *sql.DB) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    if r.Method != "GET" {
      w.Write([]byte("<b>Bad Method</b>"))
      return
    }
    my_url := r.URL.Path
    password, username, is_valid := URLToCredentials(my_url)
    if !is_valid {
      w.Write([]byte("<b>Bad URL</b>"))
      return
    }
    is_valid = EvaluatePassword(&password, &username, db)
    if !is_valid {
      w.Write([]byte("<b>Not allowed to be here</b>"))
      return
    }
    rotated_link, err := CredentialsToURL(password, &username, db)
    if err != nil {
      w.Write([]byte("<b>Something went wrong</b>"))
      return
    }
    choose_struct := ChooseStruct{NextURL: rotated_link}
    templates.ExecuteTemplate(w, "choose.html", choose_struct)
  }
}

type TrackStruct struct {
  Year int32
  Leap bool
  Days [31]int
  NextURL string
}

func TrackHandler(db *sql.DB) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    if r.Method != "GET" {
      w.Write([]byte("<b>Bad Method</b>"))
      return
    }
    my_url := r.URL.Path
    password, username, is_valid := URLToCredentials(my_url)
    if !is_valid {
      w.Write([]byte("<b>Bad URL</b>"))
      return
    }
    is_valid = EvaluatePassword(&password, &username, db)
    if !is_valid {
      w.Write([]byte("<b>Not allowed to be here</b>"))
      return
    }
    rotated_link, err := CredentialsToURL(password, &username, db)
    if err != nil {
      w.Write([]byte("<b>Something went wrong</b>"))
      return
    }
    cur_year := r.FormValue("year")
    is_valid, rtn_resp := EvaluateYear(cur_year, &rotated_link)
    if !is_valid {
      w.Write([]byte(rtn_resp))
      return
    }
    Cur_Year := StringToInt32(cur_year)
    leap_vl := is_leap(&Cur_Year)
    var track_struct TrackStruct
    track_struct = TrackStruct{Year: Cur_Year,
                             Leap: leap_vl,
                             Days: [31]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10,
                                           11, 12, 13, 14, 15, 16, 17, 18,
                                           19, 20, 21, 22, 23, 24, 25, 26, 
                                           27, 28, 29, 30, 31},
                             NextURL: rotated_link}
    err = templates.ExecuteTemplate(w, "track_page.html", track_struct)
    if err != nil {
      w.Write([]byte("<b>Something went wrong</b>"))
      return
    }
  }
}

type SeeDayStruct struct{
  NextURL string
  WorkHours int
  WorkMinutes int
  Date string
  Provided bool
  Year string
}

func SeeDayHandler(db *sql.DB) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    if r.Method != "GET" {
      w.Write([]byte("<b>Bad Method</b>"))
      return
    }
    my_url := r.URL.Path
    password, username, is_valid := URLToCredentials(my_url)
    if !is_valid {
      w.Write([]byte("<b>Bad URL</b>"))
      return
    }
    is_valid = EvaluatePassword(&password, &username, db)
    if !is_valid {
      w.Write([]byte("<b>Not allowed to be here</b>"))
      return
    }
    rotated_link, err := CredentialsToURL(password, &username, db)
    if err != nil {
      w.Write([]byte("<b>Something went wrong</b>"))
      return
    }
    cur_date := ""
    i := 9
    for my_url[i] != '_' {
      cur_date += string(my_url[i])
      i++
    }
    cur_year := ""
    i = 14
    for my_url[i] != '_' {
      cur_year += string(my_url[i])
      i++
    }
    if cur_year[0] == '-' {
      cur_year = cur_year[1:]
    }
    var content_str string
    content := db.QueryRow("SELECT name FROM " + username + " WHERE name=?;", cur_date)
    err = content.Scan(&content_str)
    var work_hours float32
    var see_day_struct SeeDayStruct
    if err != nil {
      see_day_struct = SeeDayStruct{NextURL: rotated_link,
                                       WorkHours: 0,
                                       WorkMinutes: 0,
                                       Date: cur_date,
                                       Provided: false,
                                       Year: cur_year}
      templates.ExecuteTemplate(w, "see_day.html", see_day_struct)
    } else {
      content = db.QueryRow("SELECT work_hours FROM " + username + " WHERE name=?;", cur_date)
      err = content.Scan(&work_hours)

      int_part := int(work_hours)
      pre_delta := (work_hours - float32(int_part)) * 60
      minutes := int(pre_delta)
      if pre_delta - float32(minutes) >= (float32(minutes) + 1) - pre_delta {
        minutes += 1
      }

      if err != nil {
        w.Write([]byte("<b>Something went wrong</b>"))
        return
      }
      see_day_struct = SeeDayStruct{NextURL: rotated_link,
                                       WorkHours: int_part,
                                       WorkMinutes: minutes,
                                       Date: cur_date,
                                       Provided: true,
                                       Year: cur_year}
      templates.ExecuteTemplate(w, "see_day.html", see_day_struct)
    }
  }
}

type ProvideStruct struct {
  NextURL string
  Date string
}

func ProvidePageHandler(db *sql.DB) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    if r.Method != "GET" {
      w.Write([]byte("<b>Bad Method</b>"))
      return
    }
    my_url := r.URL.Path
    password, username, is_valid := URLToCredentials(my_url)
    if !is_valid {
      w.Write([]byte("<b>Bad URL</b>"))
      return
    }
    is_valid = EvaluatePassword(&password, &username, db)
    if !is_valid {
      w.Write([]byte("<b>Not allowed to be here</b>"))
      return
    }
    rotated_link, err := CredentialsToURL(password, &username, db)
    if err != nil {
      w.Write([]byte("<b>Something went wrong</b>"))
      return
    }
    i := 14
    cur_date := ""
    for my_url[i] != '_' {
      cur_date += string(my_url[i])
      i++
    }
    provide_struct := ProvideStruct{NextURL: rotated_link,
                                    Date: cur_date}
    templates.ExecuteTemplate(w, "provide_page.html", provide_struct)
  }
}

func ProcessProvideHandler(db *sql.DB) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    if r.Method != "POST" {
      w.Write([]byte("<b>Bad Method</b>"))
      return
    }
    my_url := r.URL.Path
    password, username, is_valid := URLToCredentials(my_url)
    if !is_valid {
      w.Write([]byte("<b>Bad URL</b>"))
      return
    }
    is_valid = EvaluatePassword(&password, &username, db)
    if !is_valid {
      w.Write([]byte("<b>Not allowed to be here</b>"))
      return
    }
    rotated_link, err := CredentialsToURL(password, &username, db)
    if err != nil {
      w.Write([]byte("<n>Something went wrong</n>"))
      return
    }
    cur_date := ""
    i := 17
    for my_url[i] != '_' {
      cur_date += string(my_url[i])
      i++
    }
    nb_seconds, is_valid, resp := DateToScds(cur_date, &rotated_link)
    if !is_valid {
      w.Write([]byte(resp))
      return
    }
    data := r.FormValue("work_hours")
    work_hours, is_valid, resp := EvaluateWorkHours(data, &rotated_link)
    
    if !is_valid {
      w.Write([]byte(resp))
      return
    }
    var str_content string
    content := db.QueryRow("SELECT name FROM " + username + " WHERE name=?;", cur_date)
    err = content.Scan(&str_content)
    if err != nil {
      _, err = db.Exec("INSERT INTO " + username + " VALUE(?, ?, ?);",
                      cur_date, nb_seconds, work_hours)
      if err != nil {
        fmt.Println(err)
        w.Write([]byte("<b>Something went wrong</b>"))
        return
      }
    } else {
      _, err = db.Exec("UPDATE " + username + " SET work_hours=? WHERE name=?;",
                      work_hours, cur_date)
      if err != nil {
        fmt.Println(err)
        w.Write([]byte("<b>Something went wrong</b>"))
        return
      }
    }
    http.Redirect(w, r, "/see_day/" + cur_date + rotated_link, http.StatusFound)
  }
}

type CalculatePageStruct struct {
  NextURL string
  Year string
}

func CalculatePageHandler(db *sql.DB) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    if r.Method != "GET" {
      w.Write([]byte("<b>Bad Method</b>"))
      return
    }
    my_url := r.URL.Path
    password, username, is_valid := URLToCredentials(my_url)
    if !is_valid {
      w.Write([]byte("<b>Bad URL</b>"))
      return
    }
    is_valid = EvaluatePassword(&password, &username, db)
    if !is_valid {
      w.Write([]byte("<b>Not allowed to be here</b>"))
      return
    }
    rotated_link, err := CredentialsToURL(password, &username, db)
    if err != nil {
      w.Write([]byte("<b>Something went wrong</b>"))
      return
    }
    cur_year := ""
    i := 16
    for my_url[i] != '_' {
      cur_year += string(my_url[i])
      i++
    }
    calc_struct := CalculatePageStruct{NextURL: rotated_link,
                                       Year: cur_year}
    templates.ExecuteTemplate(w, "calculate.html", calc_struct)
  }
}

func CalculateHandler(db *sql.DB) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    if r.Method != "POST" {
      w.Write([]byte("<b>Bad Method</b>"))
      return
    }
    my_url := r.URL.Path
    password, username, is_valid := URLToCredentials(my_url)
    if !is_valid {
      w.Write([]byte("<b>Bad URL</b>"))
      return
    }
    is_valid = EvaluatePassword(&password, &username, db)
    if !is_valid {
      w.Write([]byte("<b>Not allowed to be here</b>"))
      return
    }
    rotated_link, err := CredentialsToURL(password, &username, db)
    if err != nil {
      w.Write([]byte("<b>Something went wrong</b>"))
      return
    }
    cur_year := ""
    i := 11
    for my_url[i] != '_' {
      cur_year += string(my_url[i])
      i++
    }
    bgn_date := r.FormValue("bgn_date")
    end_date := r.FormValue("end_date")
    
    bgn_scds, is_valid, resp := DateToScds(bgn_date, &rotated_link)
    if !is_valid {
      w.Write([]byte(resp))
      return
    }
    end_scds, is_valid, resp := DateToScds(end_date, &rotated_link)
    if !is_valid {
      w.Write([]byte(resp))
      return
    }
    if end_scds <= bgn_scds {
      w.Write([]byte("The begin date must be strictly ower to the end date"))
      return
    }
    var rslt string
    content := db.QueryRow("SELECT SUM(work_hours) FROM " + username + " WHERE nb_seconds BETWEEN ? AND ?;", bgn_scds, end_scds)
    err = content.Scan(&rslt)
   
    rslt_int := ""
    i = 0
    for i < len(rslt) && rslt[i] != '.' {
      rslt_int += string(rslt[i])
      i++
    }
    i++
    rslt_float := ""
    for i < len(rslt) {
      rslt_float += string(rslt[i])
      i++
    }
    if len(rslt_float) > 1 {
      rslt_float = rslt_float[:2]
      int_rslt_float := StringToInt32(rslt_float)
      int_rslt_float *= 60
      rslt_float = Int32ToString(&int_rslt_float)
      rslt_float = rslt_float[:2]
    } else if len(rslt_float) > 0 {
      rslt_float = rslt_float[:1]
      int_rslt_float := StringToInt32(rslt_float)
      int_rslt_float *= 60
      rslt_float = Int32ToString(&int_rslt_float)
      if len(rslt_float) > 1 {
        rslt_float = rslt_float[:2]
      } else {
        rslt_float = rslt_float[:2]
      }
    }

    if err != nil {
      http.Redirect(w, r, "/results/" + "0_" + cur_year + rotated_link, http.StatusFound)
    } else {
      http.Redirect(w, r, "/results/" + rslt_int + "H" + rslt_float + "_" + cur_year + rotated_link, http.StatusFound)
    }
  }
}

type ResultsStruct struct {
  NextURL string
  Result string
  Year string
}

func ResultsHandler(db *sql.DB) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    if r.Method != "GET" {
      w.Write([]byte("<b>Bad Method</b>"))
      return
    }
    my_url := r.URL.Path
    password, username, is_valid := URLToCredentials(my_url)
    if !is_valid {
      w.Write([]byte("<b>Bad URL</b>"))
      return
    }
    is_valid = EvaluatePassword(&password, &username, db)
    if !is_valid {
      w.Write([]byte("<b>Not alowed to be here</b>"))
      return
    }
    rotated_link, err := CredentialsToURL(password, &username, db)
    if err != nil {
      w.Write([]byte("<b>Something went wrong</b>"))
      return
    }
    i := 9
    rslt := ""
    for my_url[i] != '_' {
      rslt += string(my_url[i])
      i++
    }
    i++
    cur_year := ""
    for my_url[i] != '_' {
      cur_year += string(my_url[i])
      i++
    }
    rslt_struct := ResultsStruct{NextURL: rotated_link,
                                 Result: rslt,
                                 Year: cur_year}
    templates.ExecuteTemplate(w, "result.html", rslt_struct)
  }
}

func main() {

  db, err := ConnectDatabase()
  if err != nil {
    fmt.Println(err)
    return
  }

  mux := http.NewServeMux()
  
  mux.HandleFunc("/", IndexHandler)
  
  mux.HandleFunc("/login/", LoginHandler)
  mux.HandleFunc("/connection/", ConnectionHandler(db))

  mux.HandleFunc("/create_account/", CreateAccountHandler)
  mux.HandleFunc("/new_account/", NewAccountHandler(db))

  mux.HandleFunc("/choose/", ChooseHandler(db))
  mux.HandleFunc("/track_page/", TrackHandler(db))

  mux.HandleFunc("/see_day/", SeeDayHandler(db))
  mux.HandleFunc("/provide_page/", ProvidePageHandler(db))
  mux.HandleFunc("/process_provide/", ProcessProvideHandler(db))

  mux.HandleFunc("/calculate_page/", CalculatePageHandler(db))
  mux.HandleFunc("/calculate/", CalculateHandler(db))
  mux.HandleFunc("/results/", ResultsHandler(db))

  mux.Handle("/static/", http.FileServer(http.Dir(".")))

  err = http.ListenAndServe("127.0.0.1:8080", mux)
  if err != nil {
    fmt.Println(err)
    return
  }
  return
}


