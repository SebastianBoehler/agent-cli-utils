const steps = [
  {x:4, title:'Start with a guess.', caption:'The white dot is our guess. Its height is the error.'},
  {x:4, title:'Which way is downhill?', caption:'The yellow slope rises to the right. Move left.'},
  {x:2.8, title:'Take one small step.', caption:'The error falls from 9 to 3.24.'},
  {x:2.08, title:'Measure the slope again.', caption:'The slope is flatter. The next step is smaller.'},
  {x:1.648, title:'Keep moving downhill.', caption:'The dot approaches the best setting: 1.'},
  {x:1.3888, title:'Smaller steps near the bottom.', caption:'Repeat: measure the slope, then move against it.'}
];
const X = x => 650 + (x-1)*170;
const Y = error => 470 - error*37;
let path='';
for(let x=-2;x<=4.01;x+=.035){
 path+=`${path?'L':'M'}${X(x)},${Y((x-1)**2)} `;
}
document.getElementById('curve').setAttribute('d',path);
function show(index){
 const step=steps[index], loss=(step.x-1)**2, slope=2*(step.x-1);
 document.getElementById('title').textContent=step.title;
 document.getElementById('caption').textContent=step.caption;
 document.getElementById('error').textContent=loss.toFixed(2);
 const dot=document.getElementById('dot');
 dot.setAttribute('cx',X(step.x));dot.setAttribute('cy',Y(loss));
 const tangent=document.getElementById('slope');
 tangent.style.visibility=index===0?'hidden':'visible';
 tangent.setAttribute('x1',X(step.x-.35));
 tangent.setAttribute('y1',Y(loss-slope*.35));
 tangent.setAttribute('x2',X(step.x+.35));
 tangent.setAttribute('y2',Y(loss+slope*.35));
}
let index=0;show(index);
const timer=setInterval(()=>{
 show(++index);
 if(index===steps.length-1)clearInterval(timer);
},10000);
