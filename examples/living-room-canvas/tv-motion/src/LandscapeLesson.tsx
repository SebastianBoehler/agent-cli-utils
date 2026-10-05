import { AbsoluteFill, Composition, useCurrentFrame } from 'remotion';
import { PerspectiveCamera, Vector3 } from 'three';
import { cameraAt, loss, progress, settingAt } from './landscape-model';
import { LandscapeScene } from './LandscapeScene';

const stages = [
  ['A position is a choice of settings.', 'In this toy model, two settings control the predictions.'],
  ['Height shows the prediction error.', 'Each choice produces predictions. Their mistakes set the height.'],
  ['The slope gives a direction.', 'The arrow points downhill from our current choice.'],
  ['Change the settings. Watch the error.', 'One small step moves the orange point lower.'],
  ['Measure. Step. Measure again.', 'The trail records the choices we have tried.'],
  ['Better settings make smaller mistakes.', 'The model improves as its prediction error falls.'],
];

export const LandscapeLesson = () => {
  const frame = useCurrentFrame(), stage = Math.min(5, Math.floor(frame / 360));
  const setting = settingAt(progress(frame));
  const view = cameraAt(frame), camera = new PerspectiveCamera(37, 1470 / 740, 0.1, 50);
  camera.position.set(...view.position);camera.lookAt(...view.target);camera.updateMatrixWorld();
  function label(position: [number, number, number], offsetX = 0) {
    const p = new Vector3(...position).project(camera);
    return {position: 'absolute' as const, left: 60 + (p.x + 1) / 2 * 1470 + offsetX,
      top: 195 + (1 - p.y) / 2 * 740, fontSize: 40, fontWeight: 550, color: '#344a53'};
  }
  return <AbsoluteFill style={{backgroundColor: '#f5f4ee', color: '#18343e', fontFamily: 'Arial, sans-serif'}}>
    <h1 style={{position: 'absolute', left: 110, top: 76, right: 100, margin: 0,
      fontSize: 82, lineHeight: 1.1, letterSpacing: -2, fontWeight: 650}}>{stages[stage][0]}</h1>
    <LandscapeScene />
    <div style={label([0.7, 0, 2.9])}>Setting 1</div>
    <div style={{...label([-3.6, 0, 0.7], -100), translate: "-100% 0"}}>Setting 2</div>
    {stage >= 1 && <div style={{position: 'absolute', left: 1550, top: 440, width: 280}}>
      <div style={{fontSize: 44}}>Error</div>
      <div style={{fontSize: 94, lineHeight: 1.3, color: '#996015', fontVariantNumeric: 'tabular-nums'}}>{loss(setting).toFixed(2)}</div>
    </div>}
    <p style={{position: 'absolute', left: 110, right: 100, bottom: 68,
      fontSize: 42, lineHeight: 1.25, margin: 0}}>{stages[stage][1]}</p>
  </AbsoluteFill>;
};

export const RemotionRoot = () => <Composition id="LossLandscape" component={LandscapeLesson}
  width={1920} height={1080} fps={30} durationInFrames={2160} />;
