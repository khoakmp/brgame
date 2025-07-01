import "./GameIntro.css"
type Param = {
  startFn: (mode: string) => void 
}

export const GameIntro = ({startFn} :Param) =>{
  return (<div className="intro-container">
    <button className="btn" onClick={e=>startFn("single")}>SINGLE</button>
    <button className="btn" onClick={e=>startFn("multi")}>MUTLI</button>
  </div>)
}

