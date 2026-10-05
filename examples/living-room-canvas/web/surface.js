// Toy model: prediction = a + b; target = 2; error = (a + b - 2)^2.
// Two different settings can give the same perfect prediction.
let surfaceFrame=0;
window.stopSurface=()=>{cancelAnimationFrame(surfaceFrame);};
window.drawSurface=highlight=>{
 window.stopSurface();const canvas=document.getElementById('surface'),ctx=canvas.getContext('2d'),start=performance.now();
 function frame(now){
  const width=canvas.clientWidth,height=canvas.clientHeight,dpr=devicePixelRatio;
  if(canvas.width!==width*dpr||canvas.height!==height*dpr){canvas.width=width*dpr;canvas.height=height*dpr;}
  ctx.setTransform(dpr,0,0,dpr,0,0);ctx.clearRect(0,0,width,height);
  const scale=Math.min(width/1200,height/540),angle=Math.min((now-start)/4000,1)*.12;
  const project=(a,b,z)=>[width*.5+(a-b)*145*scale*Math.cos(angle),height*.64+(a+b-2)*62*scale-z*26*scale];
  function line(points,color,lineWidth=2){ctx.beginPath();points.forEach((p,i)=>{const q=project(...p);i?ctx.lineTo(...q):ctx.moveTo(...q);});ctx.strokeStyle=color;ctx.lineWidth=lineWidth;ctx.stroke();}
  // Surface mesh. The visible height is always the computed squared error.
  for(let a=0;a<=2.001;a+=.1){const points=[];for(let b=0;b<=2.001;b+=.05)points.push([a,b,(a+b-2)**2]);line(points,'#759ce8');}
  for(let b=0;b<=2.001;b+=.1){const points=[];for(let a=0;a<=2.001;a+=.05)points.push([a,b,(a+b-2)**2]);line(points,'#759ce8');}
  line([[0,0,0],[2.45,0,0]],'#dbe6f6',3);line([[0,0,0],[0,2.45,0]],'#dbe6f6',3);line([[0,0,0],[0,0,5]],'#dbe6f6',3);
  ctx.font=`${Math.max(32,38*scale)}px system-ui`;ctx.fillStyle='#dbe6f6';
  function label(a,b,z,value,dx=0,dy=0){const p=project(a,b,z);ctx.fillText(value,p[0]+dx,p[1]+dy);}
  label(2.35,0,0,'Setting 1',15,40);label(0,2.35,0,'Setting 2',-170,40);label(0,0,5,'Error',-40,-20);
  if(highlight){const valley=[];for(let a=0;a<=2;a+=.02)valley.push([a,2-a,0]);line(valley,'#7ce7bd',9);const t=Math.min((now-start)/6000,1),a=.2+1.6*t,b=.2;const p=project(a,b,(a+b-2)**2);ctx.beginPath();ctx.arc(...p,13,0,Math.PI*2);ctx.fillStyle='white';ctx.fill();}
  if(now-start<7000)surfaceFrame=requestAnimationFrame(frame);
 }
 surfaceFrame=requestAnimationFrame(frame);
};
