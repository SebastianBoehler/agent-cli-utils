import { createRoot } from 'react-dom/client';
import { Player } from '@remotion/player';
import { LandscapeLesson } from './LandscapeLesson';

const container = document.getElementById('player');
if (!container) throw new Error('The TV display container is missing.');
createRoot(container).render(<Player component={LandscapeLesson}
  durationInFrames={2160} fps={30} compositionWidth={1920} compositionHeight={1080}
  autoPlay controls={false} loop={false} clickToPlay={false} doubleClickToFullscreen={false}
  moveToBeginningWhenEnded={false} style={{width: '100vw', height: '100vh'}}
  errorFallback={({error}) => <div style={{padding: 80, fontSize: 44, color: '#a22'}}>
    The visual could not play: {error.message}
  </div>} />);
