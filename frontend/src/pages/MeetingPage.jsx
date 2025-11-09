import { useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useAppContext } from '../context/AppContext';
import { useWebRTC } from '../hooks/useWebRTC';
import { VideoTile } from '../components/VideoTile';

export function MeetingPage() {
  const { displayName, activeMeeting } = useAppContext();
  const { meetingId } = useParams();
  const navigate = useNavigate();

  useEffect(() => {
    if (!displayName) {
      navigate('/');
    }
  }, [displayName, navigate]);

  const { localStream, remoteStreams, audioEnabled, videoEnabled, toggleAudio, toggleVideo, error } = useWebRTC({
    meetingId,
    displayName,
  });

  function handleLeave() {
    navigate('/home');
  }

  return (
    <div className="meeting-layout">
      <header className="meeting-header">
        <div>
          <h2>Meeting {meetingId}</h2>
          {activeMeeting?.password && <p style={{ margin: 0, color: 'rgba(148,163,184,0.7)' }}>Passcode {activeMeeting.password}</p>}
        </div>
        <div className="controls-bar">
          <button onClick={toggleAudio}>{audioEnabled ? 'Mute' : 'Unmute'}</button>
          <button onClick={toggleVideo}>{videoEnabled ? 'Turn camera off' : 'Turn camera on'}</button>
          <button className="danger" onClick={handleLeave}>
            Leave
          </button>
        </div>
      </header>
      {error && <div className="error-message">{error}</div>}
      <section className="video-grid">
        {localStream && <VideoTile stream={localStream} label="You" mirrored />}
        {remoteStreams.map((stream) => (
          <VideoTile key={stream.id} stream={stream} label="Guest" />
        ))}
      </section>
    </div>
  );
}
