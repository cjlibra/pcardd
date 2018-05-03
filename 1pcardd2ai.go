package main  
import (  
    "crypto/md5"
    "fmt"  
    "net"  
  //  "log"  
    "os"  
	"time"
	"encoding/json"
	"strings"
	"encoding/hex" 
	"io"
	"github.com/golang/glog"
	"flag"
	
)  
const (
    HOSTIPPORT = "0.0.0.0:9900"
	AIHOSTIPPORT = "0.0.0.0:9998"
	SKEY1AI = "qiansi_ai_1"
	SKEY2AI = "qiansi_ai_2"
	skey1 = "91ylordai"
	skey2 = "91ylordai2"


)
var linkornot int
func ai2server(){
    netListen, err := net.Listen("tcp", AIHOSTIPPORT)  
    CheckError(err)  
    defer netListen.Close()  
    
    Log("Waiting for clients from ai")  
	i_ai = 0
    for {  
        conn, err := netListen.Accept()  
        if err != nil {  
            continue  
        }  
        if i_ai != 0 {
		   conn.Write([]byte("已经有一个连接，请退出后再连"))
		   Log("ai已经有一个连接，请退出后再连")
		   conn.Close()
		   
		   continue
		}
        Log(conn.RemoteAddr().String(), " tcp connect success from ai")  
		
		Conn2 = conn 
		linkornot = 0
        go aihandleConnection(conn)  
    }  



}

func aihandleConnection(conn net.Conn) {  
    defer func(){
		conn.Close()
		i_ai = 0
	}()
	i_ai = 1
    Log(conn.RemoteAddr().String(),"from ai")  
    n,err := Send_auth_req(conn ,SKEY1AI)
	if err != nil || n < 0{
	    Log(conn.RemoteAddr().String(),"send auth req error number from ai:",n,"error:",err)
		return
	}
	n , err = Read_auth_res(conn ,SKEY2AI ,SKEY1AI)
	if !(n==0 && err == nil) {
	   Log(conn.RemoteAddr().String(),"read auth res error number from ai:",n,"error:",err)
	   Send_auth_succ(conn , 0,SKEY1AI)
	   return
	}
	n , err = Send_auth_succ(conn , 1 ,SKEY1AI)
	if err != nil {
	    Log(conn.RemoteAddr().String(),"send auth succ error number from ai:",n,"error:",err)
	    return
	}
	

	for {
	    if Conn1 == nil {
		   time.Sleep(time.Second*10)
		   Log(" ai Conn1 == nil")
		   continue
		}
	    if linkornot == 1 {
		   Log("ai stop connection")
		   return
		
		}
	    n , err  = exchangesocket(conn,Conn1)
		if err != nil {
		
		     Log(conn.RemoteAddr().String(), "exchange error  from ai: ", err)  
			 if err == io.EOF || n == -99 {
			    linkornot = 1
				Log(" ai read io.EOF or n== -99,connections failed" , n)
			    return
			 }
			 time.Sleep(time.Second*2)
            // return   
		
		
		}
	
	
	
	
	}
	
	
	/*
	n,err = Read_robot_req(conn )
	if !(n==0 && err == nil) {
	     Log(conn.RemoteAddr().String(),"read robot req  error number:",n,"error:",err)
		 return
	}
	
	n,err = Send_robot_res(conn)
	if  err != nil {
	    Log(conn.RemoteAddr().String(),"send robot res  error number:",n,"error:",err)
		return
	
	}
	for {
		n,err = Read_action_cmd(conn)
		if !(n==0 && err == nil){
			Log(conn.RemoteAddr().String(),"Read_action_cmd  error number:",n,"error:",err)
			return
			
		}
	}
	*/
	
}

func fixCrcOfEx(buffer []byte ,n int, readkey string , writekey string) ([]string , int){
     type TYPETIMECRC struct {
	    Type string `json:"type" bson:"type"`
	    Time int64 `json:"time" bson:"time"`
	    Crc string `json:"crc" bson:"crc"`
	 
	 }
	 var outstring  []string
	 var typetimecrc TYPETIMECRC
	 buffer_str := StripHttpStr(string(buffer))
	 buffer_strs :=strings.Replace(buffer_str,"}{","}\x00{",-1)
	 buffer_strs =strings.Replace(buffer_strs,"}\x0a{","}\x00{",-1)
	 b_strs :=strings.Split(buffer_strs,"\x00")
	 for _,b_str := range b_strs {
		 err := json.Unmarshal([]byte(b_str) , &typetimecrc)
		 if err != nil {
		   Log(b_str,err)
		   Log(hex.EncodeToString([]byte(b_str)))
		   Log(buffer_str)
		   Log(hex.EncodeToString([]byte(buffer_str)))
		   return   outstring , -1 
		 
		 }
		 inhash := typetimecrc.Type+fmt.Sprintf("%d",typetimecrc.Time)+ readkey
		 if  typetimecrc.Crc !=  CalcMd5(inhash) {
		    Log(b_str)
			return   outstring , -2 
		 }
		 outhash := typetimecrc.Type+fmt.Sprintf("%d",typetimecrc.Time)+ writekey
		 outstring = append(outstring, strings.Replace(b_str,typetimecrc.Crc,CalcMd5(outhash),-1))
	 }
	 return  outstring , 0
	 



}

func exchangesocket(conn1 net.Conn,conn2 net.Conn)(int , error){
    var  xbuffer1 []string
    buffer := make([]byte, 20480)
    conn1.SetReadDeadline(time.Now().Add(time.Duration(2000) * time.Second))  
    
	n, err := conn1.Read(buffer) 
	Log(string(buffer))
    if err != nil {  
        Log(conn1.RemoteAddr().String(), "read error1: ", err)  
        return  -99, err
    }
	if conn1 == Conn2 {
	   xbuffer , ret := fixCrcOfEx(buffer,n,SKEY1AI,skey1)
	   if ret != 0 {
	       
	     myerr := fmt.Errorf("%s ,ret=%d", "fixCrcOfEx" ,ret)
	     return -1, myerr
	   }
	   xbuffer1  = xbuffer
	}else{
	    xbuffer , ret := fixCrcOfEx(buffer,n,skey1,SKEY1AI)
	   if ret != 0 {
	      
	     myerr := fmt.Errorf("%s ,ret=%d", "fixCrcOfEx" ,ret)
	     return  -1, myerr
	   }
	    xbuffer1  = xbuffer
	
	}
	
	for _,xbuf := range xbuffer1 {
	    if conn1 == Conn2 {
		   conn2 = Conn1
		}else{
		   conn2 = Conn2
		}
	    n , err = conn2.Write([]byte(xbuf))
	    time.Sleep(time. Millisecond * 100)
	    Log(xbuf,n)
	
	
		if err != nil {  
			Log(conn2.RemoteAddr().String(), " write error1: ", err)  
			return  -99, err
		}
	}
	
	
/*	conn2.SetReadDeadline(time.Now().Add(time.Duration(20) * time.Second))  	
	n, err = conn2.Read(buffer) 
    if err != nil {  
        Log(conn2.RemoteAddr().String(), "read error2: ", err)  
        return  n, err
    }
	if conn1 == Conn2 {
	   xbuffer , ret := fixCrcOfEx(buffer[:n],skey1 ,SKEY1AI)
	   if ret == 0 {
	      buffer = xbuffer
	   }else{
	     myerr := fmt.Errorf("%s ,ret=%d", "fixCrcOfEx" ,ret)
	     return -1, myerr
	   }
	}else{
	    xbuffer , ret := fixCrcOfEx(buffer[:n],SKEY1AI,skey1)
	   if ret == 0 {
	      buffer = xbuffer
	   }else{
	      myerr := fmt.Errorf("%s ,ret=%d", "fixCrcOfEx" ,ret)
	     return -1, myerr
	   }
	
	}
    n , err = conn1.Write(buffer[:n])
	if err != nil {  
        Log(conn1.RemoteAddr().String(), " write error2: ", err)  
        return  n, err
    }
	
	*/
	myerr := fmt.Errorf("%s" ,"success")
	return 0 ,myerr



}
var Conn1 net.Conn 
var Conn2 net.Conn 
var i_ai int
var i_fuyun int
func main() {  
  
//建立socket，监听端口
	defer func(){
	    glog.Flush()
	}()
    flag.Parse() 
	
    go ai2server()
    netListen, err := net.Listen("tcp", HOSTIPPORT)  
    CheckError(err)  
    defer netListen.Close()  
  
    Log("Waiting for clients")  
	i_fuyun = 0
    for {  
        conn, err := netListen.Accept()  
        if err != nil {  
            continue  
        }  
        if i_fuyun != 0 {
		   conn.Write([]byte("已经有一个连接，请退出后再连"))
		   Log("fuyun已经有一个连接，请退出后再连")
		   time.Sleep(time.Second*60)
		   conn.Close()
		   
		   continue
		}
        Log(conn.RemoteAddr().String(), " tcp connect success")  
		
		 
		Conn1 = conn
		linkornot = 0
        go handleConnection(conn)  
    }  
}  
//处理连接  
func handleConnection(conn net.Conn) {  
    defer func(){
	   conn.Close()
	   i_fuyun = 0
	}()
	i_fuyun = 1
    Log(conn.RemoteAddr().String())  
    n,err := Send_auth_req(conn ,skey1)
	if err != nil || n < 0{
	    Log(conn.RemoteAddr().String(),"send auth req error number:",n,"error:",err)
		return
	}
	n , err = Read_auth_res(conn ,skey2,skey1)
	if !(n==0 && err == nil) {
	   Log(conn.RemoteAddr().String(),"read auth res error number:",n,"error:",err)
	   Send_auth_succ(conn , 0 , skey1)
	   return
	}
	n , err = Send_auth_succ(conn , 1,skey1)
	if err != nil {
	    Log(conn.RemoteAddr().String(),"send auth succ error number:",n,"error:",err)
	    return
	}
	
	
	
	for {
	
	    if Conn2 == nil {
		   time.Sleep(time.Second*10)
		   Log(" fuyun Conn2 == nil")
		   continue
		}
		if linkornot == 1 {
		   Log("fuyun stop connection")
		   return
		
		}
		
	    n , err  = exchangesocket(conn,Conn2)
		if err != nil {
		
		     Log(conn.RemoteAddr().String(), "exchange error: ", err)  
			 if err == io.EOF || n == -99{
			    linkornot = 1
				Log("fuyun read io.EOF or n==-99,connections failed",n)
			    return
			 }
			 time.Sleep(time.Second*2)
            // return   
		
		
		}
	
	
	
	
	}
	
	/*
	
	n,err = Read_robot_req(conn )
	if !(n==0 && err == nil) {
	     Log(conn.RemoteAddr().String(),"read robot req  error number:",n,"error:",err)
		 return
	}
	
	n,err = Send_robot_res(conn)
	if  err != nil {
	    Log(conn.RemoteAddr().String(),"send robot res  error number:",n,"error:",err)
		return
	
	}*/
	
	/*
	for {
		n,err = Read_action_cmd(conn)
		if !(n==0 && err == nil){
			Log(conn.RemoteAddr().String(),"Read_action_cmd  error number:",n,"error:",err)
			return
			
		}
	}
    */ 
  
} 
func  read_robot_abort(buffer []byte) int {
    type ROBOTABORT struct {
	   Type string `json:"type"`
	   Robot_id int64 `json:"robot_id"`
	   Time int64 `json:"time"`
	   Crc string `json:"crc"`
	}
	var robotabort ROBOTABORT
	buffer = []byte(StripHttpStr(string(buffer)))
	json.Unmarshal([]byte(buffer),&robotabort)
	Log(robotabort)
	return 0
}
func  read_game_begin(buffer []byte) int {
    type GAMEBEGIN struct {
	    Type string `json:"type"`
	    Robot_id int64 `json:"robot_id"`
	    Yourseat int8 `json:"yourseat"`
	    Crc string `json:"crc"`
	}
	var gamebegin GAMEBEGIN
	buffer = []byte(StripHttpStr(string(buffer)))
	json.Unmarshal([]byte(buffer),&gamebegin)
	Log(gamebegin)
	return 0
}
			
func  read_play_info(buffer []byte ) int {
    type PLAYINFO struct {
	    Type string `json:"type"`
	    Robot_id int64 `json:"robot_id"`
	    Info string `json:"info"`
		Time int64 `json:"time"`
	    Crc string `json:"crc"`
	}
	var  playinfo PLAYINFO
	buffer = []byte(StripHttpStr(string(buffer)))
	json.Unmarshal([]byte(buffer),&playinfo)
	Log(playinfo)
	return 0
}
		
func  read_deal_card(buffer []byte ) int {
    type DEALCARD struct {
	    Type string `json:"type"`
	    Robot_id int64 `json:"robot_id"`
	    Cards string `json:"cards"`
		Time int64 `json:"time"`
	    Crc string `json:"crc"`
	}
	var dealcard DEALCARD
	buffer = []byte(StripHttpStr(string(buffer)))
	json.Unmarshal([]byte(buffer),&dealcard)
	Log(dealcard)
	return 0
}
			
func  read_turn(buffer []byte) int {
    type TURN struct {
	    Type	string `json:"type"`
		robot_id	int64  `json:"robot_id"`
		Turn_type	string  `json:"turn_type"`
		Seat	int8 `json:"seat"`
		Time_out	int `json:"time_out"`
		Time	int64 `json:"time"`
		Crc	 string `json:"crc"`
	
	}
	var turnturn TURN
	buffer = []byte(StripHttpStr(string(buffer)))
	json.Unmarshal([]byte(buffer),&turnturn)
	Log(turnturn)
	return 0
}
			
func  read_bid_reply(buffer []byte) int {
    type BIDREPLY struct {
	    Type	string  `json:"type"`
		Robot_id	int64  `json:"robot_id"`
		Seat	int8  `json:"seat"`
		Score	int8 `json:"score"`
		Time	int64 `json:"time"`
		Crc	string  `json:"crc"`
	
	
	}
	var bidreply BIDREPLY
	buffer = []byte(StripHttpStr(string(buffer)))
	json.Unmarshal([]byte(buffer),&bidreply)
	Log(bidreply)
	return 0
}
			
func read_bid_bottom(buffer []byte) int {
    type BIDBOTTOM struct {
	    Type	string  `json:"type"`
		Robot_id	int64 `json:"robot_id"`
		Banker	int8 `json:"banker"`
		Score	int8 `json:"score"`
		Cards	string `json:"cards"`
		Time	int64 `json:"time"`
		Crc	  string   `json:"crc"`
	
	
	}
	var  bidbottom BIDBOTTOM
	buffer = []byte(StripHttpStr(string(buffer)))
	json.Unmarshal([]byte(buffer),&bidbottom)
	Log(bidbottom)
	return 0
	
}
			
func  read_out_reply(buffer []byte ) int {
    type OUTREPLY struct {
	    Type	string  `json:"type"`
		Robot_id	int64  `json:"robot_id"`
		Seat	int8  `json:"seat"`
		Card	string  `json:"card"`
		Time	int64  `json:"time"`
		Crc	string   `json:"crc"`
	
	}
	var outreply OUTREPLY
	buffer = []byte(StripHttpStr(string(buffer)))
	json.Unmarshal([]byte(buffer),&outreply)
	Log(outreply)
	return 0
}
 			
func  read_game_end(buffer []byte) int {
    type GAMEEND struct {
	    Type	string  `json:"type"`
		Robot_id	int64  `json:"robot_id"`
		Winner	int8  `json:"winner"`
		Banker	int8 `json:"banker"`
		Score	int  `json:"score"`
		Time	int64 `json:"time"`
		Crc	string  `json:"crc"`
	}
	var gameend GAMEEND
	buffer = []byte(StripHttpStr(string(buffer)))
	json.Unmarshal([]byte(buffer),&gameend)
	Log(gameend)
	return 0
}

func  read_result(buffer []byte) int {
     type RESULTRESULT struct {
	    Seat	int8 `json:"seat"`
		Result	int8 `json:"result"`
		Score	int `json:score"`
	 
	 }
     type RESULT struct {
	    Type	string  `json:"type"`
		Robot_id	int64 `json:"robot_id"`
		Result	RESULTRESULT `json:"result"`
		Time	int64 `json:"time"`
		Crc	string  `json:"crc"`
	 }
	 var result RESULT
	 buffer = []byte(StripHttpStr(string(buffer)))
	 json.Unmarshal([]byte(buffer),&result)
	 Log(result)
	 return 0


}
func Read_action_cmd(conn net.Conn)(int,error){
    type ROBOTREQ struct {
	   Type string `json:"type"`
	   Time int64  `json:"time"`
	   Crc  string `json:"crc"`
	
	}
	buffer := make([]byte, 20480) 
	conn.SetReadDeadline(time.Now().Add(time.Duration(20) * time.Second))  
	n, err := conn.Read(buffer) 
    if err != nil {  
        Log(conn.RemoteAddr().String(), " connection error: ", err)  
        return  n, err
    }
    Log("buffer is ：",string(buffer))
    Log("hex is :", hex.EncodeToString(buffer[:n]))
	buffer_str :=  strings.TrimRight(string(buffer[:n]),"\x00")
	aa := strings.Split(buffer_str,"\x00")
	Log("len is aa",len(aa))
    for  _,actionsone := range (aa) {
		buffer = []byte(strings.TrimSpace(actionsone))
		Log("bb is:",string(buffer))
		var action_cmd ROBOTREQ

		err = json.Unmarshal(buffer,&action_cmd)
		if err != nil {
		   return -1,err

		}
		switch action_cmd.Type  {
			case "robot_abort" :
				read_robot_abort(buffer)

			case "game_begin" :
				read_game_begin(buffer)
				
			case "play_info" :
				read_play_info(buffer)
			
			case "deal_card" :
				read_deal_card(buffer)
				
			case "turn" :
				read_turn(buffer)
				
			case "bid_reply" :
				read_bid_reply(buffer)
				n,err = Send_bid_req(conn)
							Log("send bid req",n,"s",err)
				if err != nil {
					 Log(conn.RemoteAddr().String(), "Send_bid_req error number: ",n,"error :", err)  
					 return  n, err
				}
				
				
			case "bid_bottom" :
				read_bid_bottom(buffer)
				
			case "out_reply"  :
				read_out_reply(buffer)
				n,err = Send_out_req(conn)
				if err != nil {
					Log(conn.RemoteAddr().String(), "Send_out_req error number: ",n,"error :", err)  
					return  n, err
				}
				
			case "game_end" :
				read_game_end(buffer)
				
			case "result" :
				read_result(buffer)
				
			




		}
	}
	
	
	return 0,nil



}

func Send_out_req(conn net.Conn)(int ,error){
    type OUTREQ struct {
	    Type string `json:"type"`
		Robot_id int64 `json:"robot_id"`
		Seat  int8  `json:"seat"`
		Card string  `json:"card"`
		Time  int64 `json:"time"`
		Crc string `json:"crc"`
	
	
	}
	var out_req OUTREQ
	out_req.Type = "out_req"
	out_req.Robot_id = 112333434343
	out_req.Seat = 1
	out_req.Card = "黑5红8黑5红8黑5红8黑5红8黑5红8黑5红8黑5红8"
	out_req.Time = time.Now().Unix()
	str := out_req.Type + fmt.Sprintf("%d",out_req.Robot_id)+fmt.Sprintf("%d",out_req.Seat)+out_req.Card+fmt.Sprintf("%d",out_req.Time)
	out_req.Crc = CalcMd5(str)
	b_out_req , _ := json.Marshal(out_req)
	n,err := conn.Write(b_out_req)
	return n , err
	


}


func Send_bid_req(conn net.Conn)(int ,error){
    type BIDREQ struct {
	    Type string `json:"type"`
		Robot_id int64 `json:"robot_id"`
		Seat  int8  `json:"seat"`
		Score int8  `json:"score"`
		Time  int64 `json:"time"`
		Crc string `json:"crc"`
	
	
	}
	var bid_req BIDREQ
	bid_req.Type = "bid_req"
	bid_req.Robot_id = 112333434343
	bid_req.Seat = 1
	bid_req.Score = 0
	bid_req.Time = time.Now().Unix()
	str := bid_req.Type + fmt.Sprintf("%d",bid_req.Robot_id)+fmt.Sprintf("%d",bid_req.Seat)+fmt.Sprintf("%d",bid_req.Score)+fmt.Sprintf("%d",bid_req.Time)
	bid_req.Crc = CalcMd5(str)
	b_bid_req , _ := json.Marshal(bid_req)
	n,err := conn.Write(b_bid_req)
	return n , err
	


}



func Send_robot_res(conn net.Conn)(int , error){
	type ROBOTRES struct {
	   Type string `json:"type"`
	   Robot_id  int64 `json:"robot_id"`
	   Time  int64  `json:"time"`
	   Crc string `json:"crc"`

	}
	var robot_res  ROBOTRES
	robot_res.Type  =  "robot_res"
	robot_res.Robot_id = 112333434343
	robot_res.Time = time.Now().Unix()
	str := robot_res.Type + fmt.Sprintf("%d",robot_res.Robot_id)+fmt.Sprintf("%d",robot_res.Time)
	robot_res.Crc = CalcMd5(str)
	b , _ :=json.Marshal(robot_res)
	n,err := conn.Write(b)

	return n , err
   
   


}
func Read_robot_req(conn net.Conn) (int, error){
    type ROBOTREQ struct {
	   Type string `json:"type"`
	   Time int64  `json:"time"`
	   Crc  string `json:"crc"`
	
	}
	buffer := make([]byte, 20480)
    conn.SetReadDeadline(time.Now().Add(time.Duration(20) * time.Second))  	
	n, err := conn.Read(buffer) 
    if err != nil {  
        Log(conn.RemoteAddr().String(), " connection error: ", err)  
        return  n, err
    }
    var robot_req ROBOTREQ
	buffer = []byte(StripHttpStr(string(buffer)))
	Log("robot_req is : ",string(buffer))
    err = json.Unmarshal(buffer,&robot_req)	
	if robot_req.Type != "robot_req" {
	   return -1 ,nil
	}
	
	
	return  0 , nil
	

}

func CalcMd5(str string) string {
	data := []byte(str)
	has := md5.Sum(data)
	md5str1 := fmt.Sprintf("%x", has) //将[]byte转成16进制
	return md5str1

}
func  Send_auth_succ(conn net.Conn , i_succ  int , key string) (int ,error){
	s  := `{"type" : "auth_succ"   ,"result" : %d , "time" : %d, "crc" : "%s"}`
	i_time := time.Now().Unix()
	hashstr := "auth_succ" +fmt.Sprintf("%d",i_time)+key
	n , err := conn.Write([]byte(fmt.Sprintf(s,i_succ,i_time,CalcMd5(hashstr))))
	return n ,err
    

}
func Read_auth_res(conn net.Conn , key string , crckey string) (int, error){
    conn.SetReadDeadline(time.Now().Add(time.Duration(20) * time.Second))  
	buffer := make([]byte, 20480) 
	n, err := conn.Read(buffer) 
	if err != nil {  
		Log(conn.RemoteAddr().String(), " connection error: ", err)  
		return  n, err
	} 
	//Log(string(buffer))	 
	 
	skey2 := key
	type AUTHRES struct {
	Type string `json:"type" bson:"type"`
	Sign string `json:"sign" bson:"sign"`
	Time int64  `json:"time" bson:"time"`
	Crc  string `json:"crc" bson:"crc"`

	}
	var auth_res AUTHRES
	buffer = []byte(StripHttpStr(string(buffer)))
	Log(string(buffer))
	//Log(hex.EncodeToString([]byte(buffer)))
	err = json.Unmarshal(buffer,&auth_res)
	if err != nil {
	return -1, err
	}

	if auth_res.Type != "auth_res" {
	 return -2, nil
	}
	if auth_res.Sign != CalcMd5("zimakaimen"+skey2) {
	 return -3, nil
	}
    hashstr := auth_res.Type+fmt.Sprintf("%d",auth_res.Time)+crckey
	if auth_res.Crc != CalcMd5(hashstr ){
	  return -4, nil
	}

	return 0 , nil
	 

}

func Send_auth_req(conn net.Conn , key string) (int ,error){
	skey1 := key
	type AUTHREQ struct {
	Type string `json:"type" bson:"type"`
	Sign string `json:"sign" bson:"sign"`
	Time int64  `json:"time" bson:"time"`
	Crc  string `json:"crc" bson:"crc"`

	}
	var  auth_req AUTHREQ
	auth_req.Type = "auth_req"
	auth_req.Sign = "zimakaimen"
	auth_req.Time = time.Now().Unix()
	str_md5 := auth_req.Type  + fmt.Sprintf("%d",auth_req.Time)+skey1
	auth_req.Crc = CalcMd5(str_md5)
	str_auth_req ,_ := json.Marshal(auth_req)
	n,err := conn.Write([]byte(str_auth_req))
	return n,err
	 

}
func Log(v ...interface{}) {  
 //   log.Println(v...) 
    glog.V(2).Infoln(v...)	
}  
  
func CheckError(err error) {  
    if err != nil {  
        fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())  
        os.Exit(1)  
    }  
} 


func StripHttpStr(httpstr string) string{
	return strings.TrimRight(httpstr,"\x00")
 


}