import { useEffect, useRef } from "react";
import "./GamePlayer.css";
import { EventType, InputPacket, MouseEventPayload } from "../../api/rtc";

type Param = {
  stream : MediaStream | null;
  inputChannel: RTCDataChannel | null;
  closeFn : () => void;
}

export const GamePlayer = ({stream, inputChannel,closeFn}: Param) =>{
  const videoRef = useRef<HTMLVideoElement | null>(null);
  
  useEffect(()=>{  
    if(videoRef && videoRef.current) {
      videoRef.current.srcObject = stream;
    }    
  },[stream]);
  
  const createMouseEventPacket = (e: React.MouseEvent , eventType: string):InputPacket => {
    const bound = e.currentTarget.getBoundingClientRect();
    const payload : MouseEventPayload = {
      h: bound.height,
      w: bound.width,
      isleft: (e.button === 0?1:0),
      x: e.clientX - bound.left,
      y: e.clientY - bound.top,
    }    
    
    
    return {
      data: JSON.stringify(payload),
      type: eventType,
    } 
  }

  const handleMouseDown = (e : React.MouseEvent<HTMLVideoElement, MouseEvent>)=>{
    if(!inputChannel) return;
    const packet: InputPacket = createMouseEventPacket(e, EventType.MOUSE_DOWN);
    inputChannel.send(JSON.stringify(packet));
  }
  
  const handleMouseUp = (e :React.MouseEvent) =>{
    if(!inputChannel) return;
    const packet: InputPacket = createMouseEventPacket(e, EventType.MOUSE_UP);
    inputChannel.send(JSON.stringify(packet));

  }
  
  const handleMouseMove = (e :React.MouseEvent) =>{
    if(!inputChannel) return;
    const packet: InputPacket = createMouseEventPacket(e, EventType.MOUSE_MOVE);
    inputChannel.send(JSON.stringify(packet));
  }
  
  return (<div>
    <div className="control-panel">
      <button className="btn" onClick={e=>closeFn()}>EXIT</button>
    </div>
    <div className="gameplay-container">
    <video 
      className="display-window"
      autoPlay
      ref = {videoRef}
      onMouseDown={handleMouseDown}
      onMouseUp={handleMouseUp}
      onMouseMove={handleMouseMove}      
    />
    </div>
  </div>)
};