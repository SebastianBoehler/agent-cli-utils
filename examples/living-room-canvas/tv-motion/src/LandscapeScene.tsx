import { useLayoutEffect, useMemo } from 'react';
import { useThree } from '@react-three/fiber';
import { ThreeCanvas } from '@remotion/three';
import { useCurrentFrame } from 'remotion';
import { DoubleSide, Quaternion, Vector3 } from 'three';
import { cameraAt, gradient, height, progress, Setting, settingAt, trajectory } from './landscape-model';
import { point, terrain } from './terrain';

function Segment({a, b, color, radius = 0.012}: {
  a: Vector3; b: Vector3; color: string; radius?: number;
}) {
  const delta = b.clone().sub(a), midpoint = a.clone().add(b).multiplyScalar(0.5);
  const rotation = new Quaternion().setFromUnitVectors(new Vector3(0, 1, 0), delta.clone().normalize());
  return <mesh position={midpoint} quaternion={rotation}>
    <cylinderGeometry args={[radius, radius, delta.length(), 10]} />
    <meshBasicMaterial color={color} />
  </mesh>;
}

function Content() {
  const frame = useCurrentFrame(), geometry = useMemo(terrain, []), {camera} = useThree();
  const focused = frame >= 720 && frame < 1080;
  const revealHeight = frame >= 360 && frame < 720;
  const patch = useMemo(() => {
    const result = geometry.clone(), indices = geometry.index!.array, positions = geometry.attributes.position;
    const selected: number[] = [];
    for (let i = 0; i < indices.length; i += 3) {
      const vertices = [indices[i], indices[i + 1], indices[i + 2]];
      const u = vertices.reduce((sum, j) => sum + positions.getX(j), 0) / 3;
      const v = vertices.reduce((sum, j) => sum + positions.getZ(j), 0) / 3;
      if (Math.hypot(u - 1.6, v - 1.1) < 0.52) selected.push(...vertices);
    }
    result.setIndex(selected);
    return result;
  }, [geometry]);
  const view = cameraAt(frame), amount = progress(frame), setting = settingAt(amount), current = point(setting);
  useLayoutEffect(() => {
    camera.position.set(...view.position);
    camera.lookAt(...view.target);
    camera.updateProjectionMatrix();
  }, [camera, frame]);
  useLayoutEffect(() => () => {geometry.dispose();patch.dispose();}, [geometry, patch]);
  const g = gradient(setting), length = Math.hypot(...g);
  const next: Setting = [setting[0] - g[0] / length * 0.55, setting[1] - g[1] / length * 0.55];
  const arrowEnd = point(next, 0.09), arrowDirection = arrowEnd.clone().sub(current).normalize();
  const arrowRotation = new Quaternion().setFromUnitVectors(new Vector3(0, 1, 0), arrowDirection);
  const visited = [...trajectory.slice(0, Math.floor(amount) + 1), setting];
  return <>
    <ambientLight intensity={0.8} />
    <directionalLight position={[-3, 7, 5]} intensity={2.0} />
    <directionalLight position={[5, 2, -4]} intensity={0.6} />
    <mesh geometry={geometry}>
      <meshStandardMaterial vertexColors roughness={0.65} metalness={0.06} side={DoubleSide}
        transparent={focused || revealHeight} opacity={focused || revealHeight ? 0.28 : 1} depthWrite={!focused && !revealHeight} />
    </mesh>
    {focused && <mesh geometry={patch}>
      <meshStandardMaterial color="#228c87" roughness={0.8} side={DoubleSide} />
    </mesh>}
    <mesh rotation={[-Math.PI / 2, 0, 0]} position={[0, -0.08, 0]}>
      <planeGeometry args={[4.5, 4.5]} />
      <meshStandardMaterial color="#e3e7e8" roughness={1} />
    </mesh>
    <Segment a={new Vector3(-2.2, 0, 2.05)} b={new Vector3(2.2, 0, 2.05)} color="#77878c" />
    <Segment a={new Vector3(-2.05, 0, -2.2)} b={new Vector3(-2.05, 0, 2.2)} color="#77878c" />
    <mesh position={current}>
      <sphereGeometry args={[0.13, 32, 24]} />
      <meshStandardMaterial color="#f49723" roughness={0.25} metalness={0.15} />
    </mesh>
    {frame >= 360 && frame < 720 && <Segment a={new Vector3(setting[0], 0, setting[1])}
      b={new Vector3(setting[0], height(setting), setting[1])} color="#cb7512" radius={0.015} />}
    {frame >= 720 && frame < 1080 && <>
      <Segment a={current} b={arrowEnd} color="#f0ae22" radius={0.026} />
      <mesh position={arrowEnd} quaternion={arrowRotation}>
        <coneGeometry args={[0.09, 0.20, 24]} />
        <meshBasicMaterial color="#d77805" />
      </mesh>
    </>}
    {frame >= 1080 && visited.slice(1).map((p, i) =>
      <Segment key={i} a={point(visited[i], 0.085)} b={point(p, 0.085)} color="#ffba44" radius={0.018} />
    )}
  </>;
}

export const LandscapeScene = () => <ThreeCanvas width={1470} height={740}
  camera={{fov: 37, near: 0.1, far: 50, position: [3.8, 8.0, 4.2]}}
  gl={{antialias: true}} style={{position: 'absolute', left: 60, top: 195}}>
  <Content />
</ThreeCanvas>;
