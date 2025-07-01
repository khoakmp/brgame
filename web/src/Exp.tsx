import { useEffect, useState } from "react"
import { encodeBase64 } from "./utils"

export const Exp = () =>{
  const [val, setVal] = useState<string>("");


  useEffect(()=>{
    document.addEventListener("keydown", function(e) {
      console.log(e.keyCode)
    })
  },[])
  return (<div></div>)  
}