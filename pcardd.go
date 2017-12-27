package main  
import (  
    "crypto/md5"
    "fmt"  
    "net"  
    "log"  
    "os"  
	"time"
	"encoding/json"
	"strings"
	"encoding/hex" 
)  
const (
    HOSTIPPORT = "0.0.0.0:9900"


)
  
func main() {  
  
//建立socket，监听端口  
    netListen, err := net.Listen("tcp", HOSTIPPORT)  
    CheckError(err)  
    defer netListen.Close()  
  
    Log("Waiting for clients")  
    for {  
        conn, err := netListen.Accept()  
        if err != nil {  
            continue  
        }  
  
        Log(conn.RemoteAddr().String(), " tcp connect success")  
		conn.SetReadDeadline(time.Now().Add(time.Duration(20) * time.Second))  
        go handleConnection(conn)  
    }  
}  
//处理连接  
func handleConnection(conn net.Conn) {  
    defer conn.Close()
    Log(conn.RemoteAddr().String())  
    n,err := Send_auth_req(conn)
	if err != nil || n < 0{
	    Log(conn.RemoteAddr().String(),"send auth req error number:",n,"error:",err)
		return
	}
	n , err = Read_auth_res(conn )
	if !(n==0 && err == nil) {
	   Log(conn.RemoteAddr().String(),"read auth res error number:",n,"error:",err)
	   Send_auth_succ(conn , 0)
	   return
	}
	n , err = Send_auth_succ(conn , 1)
	if err != nil {
	    Log(conn.RemoteAddr().String(),"send auth succ error number:",n,"error:",err)
	    return
	}
	
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
	buffer := make([]byte, 2048) 
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
	buffer := make([]byte, 2048) 
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
func  Send_auth_succ(conn net.Conn , i_succ  int) (int ,error){
	s  := `{"type" : "auth_succ"   ,"result" : %d , "time" : %d, "crc" : "%s"}`
	i_time := time.Now().Unix()
	hashstr := "auth_succ" + fmt.Sprintf("%d",i_succ)+fmt.Sprintf("%d",i_time)
	n , err := conn.Write([]byte(fmt.Sprintf(s,i_succ,i_time,CalcMd5(hashstr))))
	return n ,err
    

}
func Read_auth_res(conn net.Conn) (int, error){
	buffer := make([]byte, 2048) 
	n, err := conn.Read(buffer) 
	if err != nil {  
		Log(conn.RemoteAddr().String(), " connection error: ", err)  
		return  n, err
	} 
	Log(string(buffer))	 
	 
	skey2 := "91ylordai2"
	type AUTHRES struct {
	Type string `json:"type" bson:"type"`
	Sign string `json:"sign" bson:"sign"`
	Time int64  `json:"time" bson:"time"`
	Crc  string `json:"crc" bson:"crc"`

	}
	var auth_res AUTHRES
	buffer = []byte(StripHttpStr(string(buffer)))
	Log(string(buffer))
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


	return 0 , nil
	 

}

func Send_auth_req(conn net.Conn) (int ,error){
	skey1 := "91ylordai"
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
	str_md5 := auth_req.Type + auth_req.Sign + fmt.Sprintf("%d",auth_req.Time)+skey1
	auth_req.Crc = CalcMd5(str_md5)
	str_auth_req ,_ := json.Marshal(auth_req)
	n,err := conn.Write([]byte(str_auth_req))
	return n,err
	 

}
func Log(v ...interface{}) {  
    log.Println(v...)  
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
