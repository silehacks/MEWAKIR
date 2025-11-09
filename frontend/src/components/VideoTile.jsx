import { useEffect, useRef } from 'react';

export function VideoTile({ stream, label, mirrored = false }) {
  const videoRef = useRef(null);

  useEffect(() => {
    if (videoRef.current && stream) {
      videoRef.current.srcObject = stream;
    }
  }, [stream]);

  return (
    <div className="video-tile">
      <video ref={videoRef} autoPlay playsInline muted={label === 'You'} style={{ transform: mirrored ? 'scaleX(-1)' : 'none' }} />
      <span className="video-label">{label}</span>
    </div>
  );
}
