"use client";

import { Canvas, useFrame } from "@react-three/fiber";
import { Float, MeshDistortMaterial, OrbitControls, Sphere } from "@react-three/drei";
import { useRef } from "react";
import type { Mesh } from "three";

function FloatingOrb({
  position,
  color,
  scale = 1,
}: {
  position: [number, number, number];
  color: string;
  scale?: number;
}) {
  const ref = useRef<Mesh>(null);
  useFrame((_, delta) => {
    if (ref.current) ref.current.rotation.y += delta * 0.3;
  });
  return (
    <Float speed={2} rotationIntensity={0.5} floatIntensity={1.5}>
      <Sphere ref={ref} args={[scale, 64, 64]} position={position}>
        <MeshDistortMaterial
          color={color}
          attach="material"
          distort={0.35}
          speed={2}
          roughness={0.2}
          metalness={0.8}
        />
      </Sphere>
    </Float>
  );
}

function Scene() {
  return (
    <>
      <ambientLight intensity={0.4} />
      <directionalLight position={[10, 10, 5]} intensity={1.2} color="#a78bfa" />
      <pointLight position={[-5, -5, -5]} intensity={0.8} color="#22d3ee" />
      <FloatingOrb position={[-2, 0.5, 0]} color="#8b5cf6" scale={1.2} />
      <FloatingOrb position={[2.2, -0.3, -1]} color="#06b6d4" scale={0.9} />
      <FloatingOrb position={[0.5, 1.2, -2]} color="#f472b6" scale={0.6} />
      <OrbitControls enableZoom={false} enablePan={false} autoRotate autoRotateSpeed={0.8} />
    </>
  );
}

export function HeroScene() {
  return (
    <div className="h-[420px] w-full rounded-3xl overflow-hidden border border-white/10 bg-gradient-to-br from-violet-950/50 to-cyan-950/30">
      <Canvas camera={{ position: [0, 0, 6], fov: 45 }}>
        <Scene />
      </Canvas>
    </div>
  );
}
