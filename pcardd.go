package main  
import (  
    "crypto/md5"
    "fmt"  
    "net"  
    "log"  
    "os"  
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
        handleConnection(conn)  
    }  
}  
//处理连接  
func handleConnection(conn net.Conn) {  
     
    Log(conn.RemoteAddr().String())  
    n,err := Send_auth_req(conn)
	if err != nil || n < 0{
	    Log(time.Now(),"--",conn.RemoteAddr().String(),"send auth req error number:",n,"error:",err)
		return
	}
	n , err = Read_auth_res(conn )
	if !(n==0 && err == nil) {
	   Log(time.Now(),"--",conn.RemoteAddr().String(),"read auth res error number:",n,"error:",err)
	   Send_auth_succ(conn , 0)
	   return
	}
	n , err = Send_auth_succ(conn , 1)
	if err != nil {
	    Log(time.Now(),"--",conn.RemoteAddr().String(),"send auth succ error number:",n,"error:",err)
	    return
	}
	
	n.err = Read_robot_req(conn )
	if !(n==0 && err == nil) {
	     Log(time.Now(),"--",conn.RemoteAddr().String(),"read robot req  error number:",n,"error:",err)
		 return
	}
	
	n,err = Send_robot_res(conn)
	if  err != nil {
	    Log(time.Now(),"--",conn.RemoteAddr().String(),"send robot res  error number:",n,"error:",err)
		return
	
	}
	for {
		n,err = Read_action_cmd(conn)
		if !(n==0 && err == nil){
			Log(time.Now(),"--",conn.RemoteAddr().String(),"Read_action_cmd  error number:",n,"error:",err)
			
		}
	}
      
  
} 
func  read_robot_abort(buffer string) int {
    type ROBOTABORT struct {
	   Type string `json:"type"`
	   Robot_id int64 `json:"robot_id"`
	   Time int64 `json:"time"`
	   Crc string `json:"crc"`
	}
}
func  read_game_begin(buffer string) int {
    type GAMEBEGIN struct {
	    Type string `json:"type"`
	    Robot_id int64 `json:"robot_id"`
	    Yourseat int8 `json:"yourseat"`
	    Crc string `json:"crc"`
	}
}
			
func  read_play_info(buffer string ) int {
    type GAMEBEGIN struct {
	    Type string `json:"type"`
	    Robot_id int64 `json:"robot_id"`
	    Info string `json:"info"`
		Time int64 `json:"time"`
	    Crc string `json:"crc"`
	}
}
		
func  read_deal_card(buffer string ) int {
    type GAMEBEGIN struct {
	    Type string `json:"type"`
	    Robot_id int64 `json:"robot_id"`
	    Cards string `json:"cards"`
		Time int64 `json:"time"`
	    Crc string `json:"crc"`
	}
}
			
func  read_turn(buffer string) int {
    type TURN struct {
	    Type	string `json:"type"`
		robot_id	int64  `json:"robot_id"`
		Turn_type	string  `json:"turn_type"`
		Seat	int8 `json:"seat"`
		Time_out	int `json:"time_out"`
		Time	int64 `json:"time"`
		Crc	 string `json:"crc"`
	
	}
}
			
func  read_bid_reply(buffer string) int {
    type BIDREPLY struct {
	    Type	string  `json:"type"`
		Robot_id	int64  `json:"robot_id"`
		Seat	int8  `json:"seat"`
		Score	int8 `json:"score"`
		Time	int64 `json:"time"`
		Crc	string  `json:"crc"`
	
	
	}
}
			
func read_bid_bottom(buffer string) int {
    type BIDBOTTOM struct {
	    Type	string  `json:"type"`
		Robot_id	int64 `json:"robot_id"`
		Banker	int8 `json:"banker"`
		Score	int8 `json:"score"`
		Cards	string `json:"cards"`
		Time	int 64 `json:"time"`
		Crc	  string   `json:"crc"`
	
	
	}
}
			
func  read_out_reply(buffer string ) int {
    type OUTREPLY struct {
	    Type	string  `json:"type"`
		Robot_id	int64  `json:"robot_id"`
		Seat	int8  `json:"seat"`
		Card	string  `json:"card"`
		Time	int64  `json:"time"`
		Crc	string   `json:"crc"`
	
	}
}
 			
func  read_game_end(buffer string) int {
    type GAMEEND struct {
	    Type	string  `json:"type"`
		Robot_id	int64  `json:"robot_id"`
		Winner	int8  `json:"winner"`
		Banker	int8 `json:"banker"`
		Score	int  `json:"score"`
		Time	int64 `json:"time"`
		Crc	string  `json:"crc"`
	}
}

func  read_result(buffer string) int {
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
    var action_cmd ROBOTREQ
	err := json.Unmarshal(buffer,&action_cmd)
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
	out_req.Robot_id = 11121313
	out_req.Seat = 1
	out_req.Card = "0"
	out_req.Time = time.Now().Unix()
	str := out_req.Type + fmt.Sprintf("%ld",out_req.Robot_id)+fmt.Sprintf("%d",out_req.Seat)+out_req.Card+fmt.Sprintf("%ld",out_req.Time)
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
	bid_req.Robot_id = 11121313
	bid_req.Seat = 1
	bid_Score = 0
	bid_req.Time = time.Now().Unix()
	str := bid_req.Type + fmt.Sprintf("%ld",bid_req.Robot_id)+fmt.Sprintf("%d",bid_req.Seat)+fmt.Sprintf("%d",bid_Score)+fmt.Sprintf("%ld",bid_req.Time)
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
   robot_res.Robot_id = 12333434343
   robot_res.Time = time.Now().Unix()
   str := robot_res.Type + fmt.Sprintf("%ld",robot_res.Robot_id)+fmt.Sprintf("%ld",robot_res.Time)
   robot_res.Crc = CalcMd5(str)
   b , _ :=json.Marshal(robot_res)
   n,err := conn.write(b)
   
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
    err = json.Unmarshal(buffer,&robot_req)	
	if robot_req.Type != "rebot_req" {
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
      s  := `{"type" : "auth_succ"   ,"result" : %d , "time" : %ld, "crc" : "%s"}`
	  i_time := time.Now().Unix()
	  hashstr := "auth_succ" + fmt.Sprintf("%d",i_succ)+fmt.Sprintf("%ld",i_time)
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
     	 
     skey2 := "91ylordai2"
	 type AUTHRES struct {
	    Type string `json: "type" bson: "type"`
		Sign string `json: "sign" bson: "sign"`
		Time int64  `json: "time" bson: "time"`
		Crc  string `json: "crc" bson: "crc"`
	 
	 }
	 var auth_res AUTHRES
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
	    Type string `json: "type" bson: "type"`
		Sign string `json: "sign" bson: "sign"`
		Time int64  `json: "time" bson: "time"`
		Crc  string `json: "crc" bson: "crc"`
	 
	 }
	 var  auth_req AUTHREQ
	 auth_req.Type = "auth_req"
	 auth_req.Sign = "zimakaimen"
	 auth_req.Time = time.Now().Unix()
	 str_md5 := auth_req.Type + auth_req.Sign + fmt.Sprintf("%ld",auth_req.Time)+skey1
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