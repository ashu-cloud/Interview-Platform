"use client";

import React, { useEffect, useRef } from "react";
import * as THREE from "three";

interface KineticSculptureProps {
  className?: string;
}

export default function KineticSculpture({ className = "w-full h-44" }: KineticSculptureProps) {
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    const width = container.clientWidth || 360;
    const height = container.clientHeight || 180;

    const scene = new THREE.Scene();
    const camera = new THREE.PerspectiveCamera(40, width / height, 0.1, 1000);
    camera.position.set(0, 0, 14);

    const renderer = new THREE.WebGLRenderer({ alpha: true, antialias: true });
    renderer.setSize(width, height);
    renderer.setPixelRatio(Math.min(window.devicePixelRatio || 1, 2));
    container.innerHTML = "";
    container.appendChild(renderer.domElement);

    const group = new THREE.Group();
    scene.add(group);

    // 1. Center faceted dual wireframe icosahedron
    const icosGeo = new THREE.IcosahedronGeometry(2.8, 1);
    const icosMat = new THREE.MeshBasicMaterial({
      color: 0x141413,
      wireframe: true,
      transparent: true,
      opacity: 0.75,
    });
    const icosMesh = new THREE.Mesh(icosGeo, icosMat);
    group.add(icosMesh);

    // 2. Inner cobalt core octahedron
    const innerGeo = new THREE.OctahedronGeometry(1.5, 0);
    const innerMat = new THREE.MeshBasicMaterial({
      color: 0x1d4ed8,
      wireframe: true,
      transparent: true,
      opacity: 0.85,
    });
    const innerMesh = new THREE.Mesh(innerGeo, innerMat);
    group.add(innerMesh);

    // 3. Acoustic concentric rings
    const ring1Geo = new THREE.RingGeometry(3.8, 3.84, 48);
    const ringMat1 = new THREE.MeshBasicMaterial({
      color: 0x141413,
      side: THREE.DoubleSide,
      transparent: true,
      opacity: 0.35,
    });
    const ring1 = new THREE.Mesh(ring1Geo, ringMat1);
    ring1.rotation.x = Math.PI / 2.3;
    group.add(ring1);

    const ring2Geo = new THREE.RingGeometry(4.7, 4.74, 48);
    const ringMat2 = new THREE.MeshBasicMaterial({
      color: 0x1d4ed8,
      side: THREE.DoubleSide,
      transparent: true,
      opacity: 0.45,
    });
    const ring2 = new THREE.Mesh(ring2Geo, ringMat2);
    ring2.rotation.y = Math.PI / 3;
    group.add(ring2);

    let mouseX = 0;
    let mouseY = 0;

    const handleMouseMove = (e: MouseEvent) => {
      const rect = container.getBoundingClientRect();
      if (rect.width && rect.height) {
        mouseX = ((e.clientX - rect.left) / rect.width - 0.5) * 1.5;
        mouseY = ((e.clientY - rect.top) / rect.height - 0.5) * 1.5;
      }
    };

    const handleResize = () => {
      if (!container) return;
      const w = container.clientWidth || 360;
      const h = container.clientHeight || 180;
      camera.aspect = w / h;
      camera.updateProjectionMatrix();
      renderer.setSize(w, h);
    };

    window.addEventListener("mousemove", handleMouseMove);
    window.addEventListener("resize", handleResize);

    const startTime = performance.now();
    let animId: number;

    const animate = () => {
      animId = requestAnimationFrame(animate);
      const t = (performance.now() - startTime) * 0.001;

      icosMesh.rotation.x = t * 0.15;
      icosMesh.rotation.y = t * 0.22;
      innerMesh.rotation.x = -t * 0.3;
      innerMesh.rotation.z = t * 0.25;
      ring1.rotation.z = t * 0.1;
      ring2.rotation.x = t * -0.12;

      group.rotation.y += (mouseX - group.rotation.y) * 0.05;
      group.rotation.x += (mouseY - group.rotation.x) * 0.05;

      renderer.render(scene, camera);
    };

    animate();

    return () => {
      cancelAnimationFrame(animId);
      window.removeEventListener("mousemove", handleMouseMove);
      window.removeEventListener("resize", handleResize);
      renderer.dispose();
      icosGeo.dispose();
      innerGeo.dispose();
      ring1Geo.dispose();
      ring2Geo.dispose();
    };
  }, []);

  return <div ref={containerRef} className={`relative overflow-hidden ${className}`} />;
}
