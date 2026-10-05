const plot=document.getElementById('plot');
const stages=[
 {title:'You choose the setting.',caption:'Our simple model predicts the number you choose.',mode:'choice'},
 {title:'We measure the mistake.',caption:'Prediction 4 − target 2 = a gap of 2.',mode:'gap'},
 {title:'Each choice gets an error height.',caption:'Height = gap × gap. A perfect prediction has height zero.',mode:'curve'},
 {title:'Two settings make a surface.',caption:'Across: two settings. Up: the error they produce.',mode:'surface'},
 {title:'Find settings in the valley.',caption:'Lower is better. We must find which choices lead there.',mode:'valley'}
];
function text(x,y,value,color='#dbe6f6',size=44){return `<text x="${x}" y="${y}" fill="${color}" style="fill:${color};font-size:${size}px">${value}</text>`;}
function bar(x,height,color){return `<rect x="${x}" y="${500-height}" width="220" height="${height}" rx="14" fill="${color}"/>`;}
function choice(){return text(135,120,'Setting',undefined,48)+text(135,360,'4','#7ce7bd',140)+text(640,320,'→','#dbe6f6',110)+text(1010,120,'Prediction',undefined,48)+text(1040,360,'4','#759ce8',140);}
function gap(){return bar(380,360,'#759ce8')+bar(990,180,'#7ce7bd')+text(370,580,'Prediction 4')+text(1000,580,'Target 2')+`<path d="M740 140H810M780 140V320M740 320H810" fill="none" stroke="#ffc878" stroke-width="8"/>`+text(825,245,'Gap 2','#ffc878',60);}
const X=w=>220+w*280,Y=e=>470-e*80;
function curve(){let d='';for(let w=0;w<=4.001;w+=.02)d+=`${d?'L':'M'}${X(w)},${Y((w-2)**2)} `;let dots='';for(const w of [0,1,2,3,4])dots+=`<circle cx="${X(w)}" cy="${Y((w-2)**2)}" r="13" fill="${w===2?'#7ce7bd':'white'}"/>`+text(X(w)-14,535,w)+text(X(w)-14,Y((w-2)**2)-35,(w-2)**2,w===2?'#7ce7bd':'#dbe6f6');return `<path d="M220 70V470H1430" fill="none" stroke="#436681" stroke-width="4"/><path d="${d}" fill="none" stroke="#759ce8" stroke-width="7"/>`+dots+text(140,55,'Error')+text(1070,600,'Setting');}
function render(i){const s=stages[i];document.getElementById('title').textContent=s.title;document.getElementById('caption').textContent=s.caption;const is3d=s.mode==='surface'||s.mode==='valley';plot.hidden=is3d;document.getElementById('surface').hidden=!is3d;if(is3d)window.drawSurface(s.mode==='valley');else{window.stopSurface();plot.innerHTML=s.mode==='choice'?choice():s.mode==='gap'?gap():curve();}}
let stage=0;render(0);const timer=setInterval(()=>{render(++stage);if(stage===stages.length-1)clearInterval(timer);},12000);
