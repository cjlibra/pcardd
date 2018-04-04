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
	
	heart_timeout = 5


)
 
func ai2server(){
    netListen, err := net.Listen("tcp", AIHOSTIPPORT)  
    CheckError(err)  
    defer netListen.Close()  
    
    glog.V(2).Infoln("Waiting for clients from ai")  
	i_ai = 0
    for {  
        conn, err := netListen.Accept()  
        if err != nil {  
            continue  
        }  
        if i_ai != 0 {
		   conn.Write([]byte("you has another connection existed ,please close that!!!---已经有一个连接，请退出后再连"))
		   glog.V(2).Infoln("ai已经有一个连接，请退出后再连")
		   conn.Close()
		   
		   continue
		}
        glog.V(2).Infoln(conn.RemoteAddr().String(), " tcp connect success from ai")  
		
		Conn2 = conn 
		 
        go aihandleConnection(conn)  
    }  



}

func aihandleConnection(conn net.Conn) {  
    defer func(){
		conn.Close()
		i_ai = 0
	}()
	i_ai = 1
    glog.V(2).Infoln(conn.RemoteAddr().String(),"from ai")  
    n,err := Send_auth_req(conn ,SKEY1AI)
	if err != nil || n < 0{
	    glog.V(2).Infoln(conn.RemoteAddr().String(),"send auth req error number from ai:",n,"error:",err)
		return
	}
	n , err = Read_auth_res(conn ,SKEY2AI ,SKEY1AI)
	if !(n==0 && err == nil) {
	   glog.V(2).Infoln(conn.RemoteAddr().String(),"read auth res error number from ai:",n,"error:",err)
	   Send_auth_succ(conn , 0,SKEY1AI)
	   return
	}
	n , err = Send_auth_succ(conn , 1 ,SKEY1AI)
	if err != nil {
	    glog.V(2).Infoln(conn.RemoteAddr().String(),"send auth succ error number from ai:",n,"error:",err)
	    return
	}
	go start_send_heart_req(conn   , SKEY1AI)

	for {
	
	   
		
		
	   
	   
	    n , err  = exchangesocket(conn,Conn1)
		if err != nil {
		
		     glog.V(2).Infoln(conn.RemoteAddr().String(), "exchange error  from ai: ", err)  
			 if  err == io.EOF || n == -98 {
			     glog.V(2).Infoln(" ai connections lost" , n)
			     return
			 }
			 if  n == -99 {
			   
				glog.V(2).Infoln(" fuyun connections lost" , n)
			    
			 }
			 time.Sleep(time.Second*2)
             
		
		
		}
	
	
	
	
	}
	
 
	
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
	 buffer_strs =strings.Replace(buffer_strs,"}\x00\x00{","}\x00{",-1)
	 buffer_strs =strings.Replace(buffer_strs,"}\x0a{","}\x00{",-1)
	 b_strs :=strings.Split(buffer_strs,"\x00")
	 for _,b_str := range b_strs {
		 err := json.Unmarshal([]byte(b_str) , &typetimecrc)
		 if err != nil {
		   glog.V(2).Infoln(b_str,err)
		   glog.V(2).Infoln(hex.EncodeToString([]byte(b_str)))
		   glog.V(2).Infoln(buffer_str)
		   glog.V(2).Infoln(hex.EncodeToString([]byte(buffer_str)))
		   return   outstring , -1 
		 
		 }
		 if typetimecrc.Type == "ai_ready_req" || typetimecrc.Type == "heart_res" {
			continue
		 }
		 inhash := typetimecrc.Type+fmt.Sprintf("%d",typetimecrc.Time)+ readkey
		 if  typetimecrc.Crc !=  CalcMd5(inhash) {
		    glog.V(2).Infoln(b_str)
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
    conn1.SetReadDeadline(time.Now().Add(time.Duration(100) * time.Second))  
    
	n, err := conn1.Read(buffer)	
    if err != nil {  
        glog.V(2).Infoln(conn1.RemoteAddr().String(), "read error1: ", err)  
        return  -98, err
    }
	glog.V(2).Infoln(string(buffer[:n]))
	var readkey  string
	if conn1 == Conn2 {
	   readkey = SKEY1AI
	}else{
	   readkey = skey1
	}
	if check_heart_res(buffer ,readkey) != 0 {
	    return  -6 , fmt.Errorf("%s : heart res error" , conn1.RemoteAddr().String())
	}
	 if conn2 == nil {
	  // glog.V(2).Info(conn1.RemoteAddr().String(),"conn2 is null")
	   if conn1 == Conn2 {
			return -5 ,fmt.Errorf("%s :fuyun conn is null" , conn1.RemoteAddr().String())
	   }else{
	        return -5 ,fmt.Errorf("%s : ai conn is null" , conn1.RemoteAddr().String())
	   }
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
	    glog.V(2).Infoln(xbuf,n)
	
	
		if err != nil {  
			glog.V(2).Infoln(conn2.RemoteAddr().String(), " write error1: ", err)  
			return  -99, err
		}
	}
	
	
	myerr := fmt.Errorf("%s" ,"success")
	return 0 ,myerr



}
var ch_heart0 chan int
var ch_heart1 chan int
func check_heart_res(buffer []byte ,readkey string)  int {
     type TYPETIMECRC struct {
	    Type string `json:"type" bson:"type"`
	    Time int64 `json:"time" bson:"time"`
	    Crc string `json:"crc" bson:"crc"`
	 
	 }
     var typetimecrc TYPETIMECRC
	 buffer_str := StripHttpStr(string(buffer))
	 buffer_strs :=strings.Replace(buffer_str,"}{","}\x00{",-1)
	 buffer_strs =strings.Replace(buffer_strs,"}\x00\x00{","}\x00{",-1)
	 buffer_strs =strings.Replace(buffer_strs,"}\x0a{","}\x00{",-1)
	 b_strs :=strings.Split(buffer_strs,"\x00")
	 for _,b_str := range b_strs {
		 err := json.Unmarshal([]byte(b_str) , &typetimecrc)
		 if err != nil {
		   glog.V(2).Infoln(b_str,err)
		   glog.V(2).Infoln(hex.EncodeToString([]byte(b_str)))
		   glog.V(2).Infoln(buffer_str)
		   glog.V(2).Infoln(hex.EncodeToString([]byte(buffer_str)))
		   return     -1 
		 
		 }
		 if typetimecrc.Type == "ai_ready_req" {
		    inhash1 := typetimecrc.Type+fmt.Sprintf("%d",typetimecrc.Time)+ readkey
			if  typetimecrc.Crc !=  CalcMd5(inhash1) {
				glog.V(2).Infoln(b_str)
			    return   -3
		    }
			type AIREADYRES struct {
				Type string `json:"type" bson:"type"`
				Result int `json:"result" bson:"result"`
				Time int64 `json:"time" bson:"time"`
				Crc string `json:"crc" bson:"crc"`
	 
			}
			var aireadyres AIREADYRES
			aireadyres.Type = "ai_ready_res"
			if ai_active == 1 {
				aireadyres.Result = 1
			}else{
			    aireadyres.Result = 0
			}
			aireadyres.Time = time.Now().Unix()
			inhash2 := aireadyres.Type+fmt.Sprintf("%d",aireadyres.Time)+ readkey
			aireadyres.Crc = CalcMd5(inhash2)
			baireadyres , err := json.Marshal(aireadyres)
			if err != nil {
			   return -4
			}
			n, err := Conn1.Write(baireadyres)
			if err != nil {
			    glog.V(2).Infoln(err)
			}else{
				glog.V(2).Infoln(string(baireadyres[:n]))
			}
		    continue
		 }
		 if typetimecrc.Type != "heart_res" {
		     continue
		 }
		 
		 inhash := typetimecrc.Type+fmt.Sprintf("%d",typetimecrc.Time)+ readkey
		 if  typetimecrc.Crc !=  CalcMd5(inhash) {
		    glog.V(2).Infoln(b_str)
			return     -2 
		 }
		 if readkey == SKEY1AI {
			ch_heart0 <- 0
		 }else{
			ch_heart1 <- 1
		 }
		 
	 }
     return 0

}

var ai_active int
func ai_check_ch_heart(){
    for {
	   select {
	       case  <-ch_heart0 :
		       time.Sleep(time.Millisecond*50)
		       ai_active = 1
	       
			   
	       case <- time.After(3*heart_timeout * time.Second):
		      ai_active = 0
		      if aiflag_start_send_heart_req == 0 {
				fmt.Println("aiflag_flag_start_send_heart_req = 0") 
				continue
			  }
		      glog.V(2).Infoln("ai heart res timeout")
		       
			  if   Conn2 != nil {
				   
				  Conn2.Close()
			  }
	   
	   
	   }
	
	
	}

}
func check_ch_heart(){
    for {
	   select {
	        
		       
	       case  <-ch_heart1 :
		       time.Sleep(time.Millisecond*50)
			   
	       case <- time.After(3*heart_timeout * time.Second):
		      if flag_start_send_heart_req == 0 {
				fmt.Println("flag_flag_start_send_heart_req = 0") 
				continue
			  }
		      glog.V(2).Infoln("fuyun heart res timeout")
		       
			  if Conn1 != nil   {
				  Conn1.Close()
				   
			  }
	   
	   
	   }
	
	
	}

}
var Conn1 net.Conn 
var Conn2 net.Conn 
var i_ai int
var i_fuyun int
func main() {  
    ch_heart0 = make(chan int ,256)
	ch_heart1 = make(chan int ,256)
	go check_ch_heart()
	go ai_check_ch_heart()
//建立socket，监听端口
	defer func(){
	    glog.Flush()
	}()
    flag.Parse() 
	
    go ai2server()
    netListen, err := net.Listen("tcp", HOSTIPPORT)  
    CheckError(err)  
    defer netListen.Close()  
  
    glog.V(2).Infoln("Waiting for clients")  
	i_fuyun = 0
    for {  
        conn, err := netListen.Accept()  
        if err != nil {  
            continue  
        }  
        if i_fuyun != 0 {
		   conn.Write([]byte("you has another connection existed ,please close that!!!---已经有一个连接，请退出后再连"))
		   glog.V(2).Infoln("fuyun已经有一个连接，请退出后再连")
		   time.Sleep(time.Second*60)
		   conn.Close()
		   
		   continue
		}
        glog.V(2).Infoln(conn.RemoteAddr().String(), " tcp connect success")  
		
		 
		Conn1 = conn
		 
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
    glog.V(2).Infoln(conn.RemoteAddr().String())  
    n,err := Send_auth_req(conn ,skey1)
	if err != nil || n < 0{
	    glog.V(2).Infoln(conn.RemoteAddr().String(),"send auth req error number:",n,"error:",err)
		return
	}
	n , err = Read_auth_res(conn ,skey2,skey1)
	if !(n==0 && err == nil) {
	   glog.V(2).Infoln(conn.RemoteAddr().String(),"read auth res error number:",n,"error:",err)
	   Send_auth_succ(conn , 0 , skey1)
	   return
	}
	n , err = Send_auth_succ(conn , 1,skey1)
	if err != nil {
	    glog.V(2).Infoln(conn.RemoteAddr().String(),"send auth succ error number:",n,"error:",err)
	    return
	}
	
	go start_send_heart_req(conn   , skey1)
	
	for {
	
	
	    
	
	    
		
		
	    n , err  = exchangesocket(conn,Conn2)
		if err != nil {
		
		     glog.V(2).Infoln(conn.RemoteAddr().String(), "exchange error: ", err,"ret is:",n)  
			 if err == io.EOF || n == -98{
			     
				glog.V(2).Infoln("fuyun read io.EOF or n==-98,connections lost",n)
			    return
			 }
			 if   n == -99{
			    glog.V(2).Infoln("ai connections lost",n)
			 }
			 time.Sleep(time.Second*2)
            // return   
		
		
		}
	
	
	
	
	}
	
	
  
} 
var aiflag_start_send_heart_req int
var flag_start_send_heart_req int
func start_send_heart_req(conn net.Conn , readkey string){
    type HEARTREQ struct {
	   Type string `json:"type"`
	   Time int64 `json:"time"`
	   Crc string `json:"crc"`
	}
	defer func(){
	    if readkey == SKEY1AI {
	        aiflag_start_send_heart_req = 0
		}else{
		    flag_start_send_heart_req = 0
		}
	}()
	if readkey == SKEY1AI {
		aiflag_start_send_heart_req = 1
	}else{
		flag_start_send_heart_req = 1
	}
	var hearreq HEARTREQ
	hearreq.Type = "heart_req"
	for {
	     
		hearreq.Time = time.Now().Unix()
		
		inhash := hearreq.Type+fmt.Sprintf("%d",hearreq.Time)+ readkey
		hearreq.Crc = CalcMd5(inhash)
		
		heartreqstr , err := json.Marshal(hearreq)
		if err != nil {
		   glog.V(2).Infoln(conn.RemoteAddr().String(),err)
		   continue
		
		}
		_, err = conn.Write([]byte(heartreqstr))
		if err != nil {
			 glog.V(2).Infoln(conn.RemoteAddr().String(),err)
			 return
		}
		time.Sleep(time.Second * heart_timeout)
	
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
	glog.V(2).Infoln(robotabort)
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
	glog.V(2).Infoln(gamebegin)
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
	glog.V(2).Infoln(playinfo)
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
	glog.V(2).Infoln(dealcard)
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
	glog.V(2).Infoln(turnturn)
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
	glog.V(2).Infoln(bidreply)
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
	glog.V(2).Infoln(bidbottom)
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
	glog.V(2).Infoln(outreply)
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
	glog.V(2).Infoln(gameend)
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
	 glog.V(2).Infoln(result)
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
        glog.V(2).Infoln(conn.RemoteAddr().String(), " connection error: ", err)  
        return  n, err
    }
    glog.V(2).Infoln("buffer is ：",string(buffer))
    glog.V(2).Infoln("hex is :", hex.EncodeToString(buffer[:n]))
	buffer_str :=  strings.TrimRight(string(buffer[:n]),"\x00")
	aa := strings.Split(buffer_str,"\x00")
	glog.V(2).Infoln("len is aa",len(aa))
    for  _,actionsone := range (aa) {
		buffer = []byte(strings.TrimSpace(actionsone))
		glog.V(2).Infoln("bb is:",string(buffer))
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
							glog.V(2).Infoln("send bid req",n,"s",err)
				if err != nil {
					 glog.V(2).Infoln(conn.RemoteAddr().String(), "Send_bid_req error number: ",n,"error :", err)  
					 return  n, err
				}
				
				
			case "bid_bottom" :
				read_bid_bottom(buffer)
				
			case "out_reply"  :
				read_out_reply(buffer)
				n,err = Send_out_req(conn)
				if err != nil {
					glog.V(2).Infoln(conn.RemoteAddr().String(), "Send_out_req error number: ",n,"error :", err)  
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
        glog.V(2).Infoln(conn.RemoteAddr().String(), " connection error: ", err)  
        return  n, err
    }
    var robot_req ROBOTREQ
	buffer = []byte(StripHttpStr(string(buffer)))
	glog.V(2).Infoln("robot_req is : ",string(buffer))
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
		glog.V(2).Infoln(conn.RemoteAddr().String(), " connection error: ", err)  
		return  n, err
	} 
	//glog.V(2).Infoln(string(buffer))	 
	 
	skey2 := key
	type AUTHRES struct {
	Type string `json:"type" bson:"type"`
	Sign string `json:"sign" bson:"sign"`
	Time int64  `json:"time" bson:"time"`
	Crc  string `json:"crc" bson:"crc"`

	}
	var auth_res AUTHRES
	buffer = []byte(StripHttpStr(string(buffer)))
	glog.V(2).Infoln(string(buffer))
	//glog.V(2).Infoln(hex.EncodeToString([]byte(buffer)))
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
   
  
func CheckError(err error) {  
    if err != nil {  
        fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())  
        os.Exit(1)  
    }  
} 


func StripHttpStr(httpstr string) string{
	return strings.TrimRight(httpstr,"\x00")
 


}
