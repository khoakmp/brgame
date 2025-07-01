import { useEffect, useRef, useState } from 'react';
import { GameIntro } from './components/GameIntro/GameIntro';
import { GamePlayer } from './components/GamePlayer/GamePlayer';
import { MessageType, RequestGamePayload, SessionCreatedPayload, WsMessage } from './api/message';
import { EventType, InputPacket, KeyEventPayload } from './api/rtc';
import { decodeBase64, encodeBase64, genString } from './utils';
import "./App.css"
const keyBitmap = new Uint32Array(8);
let pendingWsMessages :WsMessage[] = [];

type GameSession  = {
  sessionID :string;
  workerID: string;
  playerIDs: string[];
}

type WaitState  = {
  iswatting: boolean;
  waitted: number;
}

function App() {
  const [pageState, setPageState] = useState<string>("init"); 
  const [ws, setWs] = useState<WebSocket | null>(null); 
  const [clientId, setClientId] = useState<string>("");
  const [peerConn, setPeerConn] = useState<RTCPeerConnection | null>(null);
  const [mediaStream, setMediaStream] = useState<MediaStream | null>(null);
  const [inputChannel, setInputChannel] = useState<RTCDataChannel | null> (null);
  const [waitState, setWaitState] = useState<boolean>(false)
  const [waitCount, setWaitCount] = useState<number>(0);
  const intervalFn = useRef<NodeJS.Timer | null>(null);
  useEffect(()=>{    
    const clientID = genString(6);
    setClientId(clientID);
  },[]);

  const handleMessageRTC = async (msg :WsMessage) => {
    if(!ws || !peerConn) return;
    switch(msg.type) {
      case MessageType.ICECandidate:
        //console.log("Recv ICECandidate:",msg.payload )
        const candidate = JSON.parse(decodeBase64(msg.payload));
        
        const ice = new RTCIceCandidate(candidate);
        try{
          await peerConn.addIceCandidate(ice);
        }catch(e){
          console.log("failed to add icecandidate:",e)
        }          
        break;
      case MessageType.SDP:
        
        //console.log("Recv SDP:", msg.payload);
        const sdp = JSON.parse(decodeBase64(msg.payload));
        const remoteSDP = new RTCSessionDescription(sdp);
        await peerConn.setRemoteDescription(remoteSDP);
        const localSDP = await peerConn.createAnswer();
        peerConn.setLocalDescription(localSDP);
      
        ws.send(JSON.stringify({
          payload: encodeBase64(JSON.stringify(localSDP)),
          receiver_ids : [msg.sender_id],
          sender_id: clientId,
          session_id: msg.session_id,
          type: MessageType.SDP,
        } as WsMessage))      
            
        break;
      case MessageType.WaitTimeout:
        setWaitState(false);
        alert("Wait Timeout")
    } 
  }
  
  useEffect(()=>{
    if(clientId.length === 0 ) return;
    console.log("connecting ws...");
    const url = "ws://localhost:8080/ws?client_id="+clientId+"&role=client";
    
    const websocket = new WebSocket(url);    
  
    setWs(websocket);
  },[clientId]);
  
  useEffect(()=>{
    if(ws ===null) return;
   
    if(peerConn === null) {      
      ws.onmessage = async (event)=>{
        const msg : WsMessage = JSON.parse(event.data);
        console.log("Recv msg type:", msg.type)
        switch(msg.type) {
          case MessageType.SessionCreated:
            const payload : SessionCreatedPayload = JSON.parse(msg.payload);
            initPeerConn(payload.session_id, payload.worker_id);
            setWaitState(false);
            setPageState("play_game"); 
            return
          case MessageType.WaitTimeout:
            alert("there are no other players")
            setWaitState(false)
            break
          case MessageType.WorkerNotFound:
            setWaitState(false)
            alert("service unavailable")
            break
          default:
            pendingWsMessages.push(msg)
        }        

      }
      return;
    }
     
    for(let i=0;i<pendingWsMessages.length;i++) {
      handleMessageRTC(pendingWsMessages[i]);;
    }
    pendingWsMessages=[];
    
    ws.onmessage = async (ev : MessageEvent) =>{
      const msg : WsMessage = JSON.parse(ev.data as string);
      console.log("recv ws message type",msg.type)
      await handleMessageRTC(msg); 
    }
  }, [peerConn,ws])
  
  useEffect(()=>{
    const handleKeyDown = (event : KeyboardEvent) => {
      if(!inputChannel || inputChannel.readyState !== "open") return;

      const keyCode = event.keyCode;
      const index =keyCode >>> 5;
      const bit = 1 << (keyCode & 31);
      
      if((keyBitmap[index] & bit) >0) {
        return;
      }
      keyBitmap[index] |= bit;
    
      const payload : KeyEventPayload = {
        keycode: keyCode
      }        
      const inputData : InputPacket = {
        data: JSON.stringify(payload),
        type: EventType.KEY_DOWN
      }      
      //console.log("Send key down event:", inputData);
      inputChannel.send(JSON.stringify(inputData));
    }        
    
    const handleKeyUp = (event: KeyboardEvent) =>{
      if(!inputChannel || inputChannel.readyState !=="open") return;
      
      const keyCode = event.keyCode;
      const index =keyCode >>> 5;
      const bit = 1 << (keyCode & 31);
      keyBitmap[index] &= ~bit;
      const payload : KeyEventPayload = {
        keycode: keyCode
      }        
      const inputData : InputPacket = {
        data: JSON.stringify(payload),
        type: EventType.KEY_UP
      }
      //console.log("Send key up event:", inputData);

      inputChannel.send(JSON.stringify(inputData));
    }    

    document.addEventListener("keydown", handleKeyDown);
    document.addEventListener("keyup", handleKeyUp);

    return ()=>{
      document.removeEventListener("keydown", handleKeyDown);
      document.removeEventListener("keyup", handleKeyUp);
    };
  },[inputChannel]);
  
  const initPeerConn = (sessionID :string, workerID :string) => {
    console.log("init peer conn");
    
    const peerConn = new RTCPeerConnection({
      iceServers: [
        {
          urls: "stun:stun.l.google.com:19302",
        },
      ],
    });
    
    peerConn.onicecandidate = (event: RTCPeerConnectionIceEvent) =>{

      const candidate = event.candidate;
      if(!candidate || !ws)  {
        return 
      }
      console.log("client ice candidate:", candidate);
      
      const msg : WsMessage = {
        type: MessageType.ICECandidate,
        payload: encodeBase64(JSON.stringify(candidate)),
        receiver_ids: [workerID],
        sender_id: clientId,
        session_id: sessionID
      }

      console.log("Send ice candidate msg:", msg);
       
      ws.send(JSON.stringify(msg))
    }
    
    peerConn.ontrack = (event: RTCTrackEvent) =>{
      if(event.streams && event.streams[0]) {
        console.log("get stream id:", event.streams[0].id);
        setMediaStream(event.streams[0]);
        return;                
      }        
      if(mediaStream) {
        const curStream = {...mediaStream} as MediaStream;
        curStream.addTrack(event.track);
        console.log("add track, label:", event.track.label, "id:" ,event.track.id);

        setMediaStream(curStream);
        return;
      } 
      const curStream = new MediaStream();
      curStream.addTrack(event.track);
      console.log("add track, label:", event.track.label, "id:" ,event.track.id);
      
      setMediaStream(curStream);
    }
    peerConn.ondatachannel= (event) =>{
      console.log("recv data channel, label:", event.channel.label, "id:", event.channel.id);
      setInputChannel(event.channel);
    }        

    peerConn.oniceconnectionstatechange = (event) =>{
          
    } 
   
    setPeerConn(peerConn);
  }
  
  const closeGame = () =>{
    inputChannel?.close();
    peerConn?.close();
    pendingWsMessages=[];
    
    setPeerConn(null);
    setMediaStream(null);
    setInputChannel(null);
    setPageState("init");          
  }
 
  const startGame = (mode: string) =>{  
    if(!ws) {
      return
    }

    const payload :RequestGamePayload = {
      app_name : "bloody_roar_2",
      mode: mode, // multi or single 
    }
    
    const msg : WsMessage = {
      payload: JSON.stringify(payload),
      receiver_ids: [],
      sender_id: clientId,
      session_id: "",
      type: MessageType.RequestGame
    }

    ws.send(JSON.stringify(msg))
    console.log("Sent game request")
    if(!waitState) {
      console.log("go here")
      setWaitState(true);
    } 
  };

  useEffect(()=>{
    console.log("hey")
    if(waitState) {
      intervalFn.current = setInterval(countWait, 1000)
      return
    }    
    if(intervalFn.current) {
      clearInterval(intervalFn.current)
    }
  },[waitState])
 
  const countWait = () =>{
    setWaitCount((waitCount)=> waitCount+1);    
  }
  
  return (
    <div className="App">
      {waitState?<div className='btn'>
        waiting...{waitCount}
      </div>:""}
      {pageState == "init"?
      <GameIntro startFn= {startGame}/>:
      <GamePlayer 
        stream={mediaStream} 
        inputChannel={inputChannel} 
        closeFn={closeGame}
      />
      }
    </div>
  );
}

export default App;
