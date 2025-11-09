import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { api } from '../hooks/useApi';
import { useAppContext } from '../context/AppContext';

export function HomePage() {
  const { displayName, activeMeeting, setActiveMeeting } = useAppContext();
  const [meetingId, setMeetingId] = useState('');
  const [password, setPassword] = useState('');
  const [feedback, setFeedback] = useState(null);
  const [error, setError] = useState(null);
  const navigate = useNavigate();

  useEffect(() => {
    if (!displayName) {
      navigate('/');
    }
  }, [displayName, navigate]);

  useEffect(() => {
    if (activeMeeting) {
      setMeetingId(activeMeeting.id || '');
    }
  }, [activeMeeting]);

  async function handleCreate() {
    setError(null);
    setFeedback(null);
    try {
      const meeting = await api.createMeeting();
      setActiveMeeting(meeting);
      setFeedback(`Meeting created! ID ${meeting.id} • Passcode ${meeting.password}`);
    } catch (err) {
      setError(err.message || 'Could not create meeting.');
    }
  }

  async function handleJoin(event) {
    event.preventDefault();
    setError(null);
    setFeedback(null);
    try {
      const meeting = await api.joinMeeting(meetingId.trim(), password.trim());
      setActiveMeeting({ ...meeting, password: password.trim() });
      navigate(`/meeting/${meeting.id}`);
    } catch (err) {
      setError(err.message || 'Unable to join meeting.');
    }
  }

  return (
    <div className="main-layout">
      <div className="card" style={{ width: 'min(520px, 100%)' }}>
        <h1>Hello, {displayName || 'Guest'}</h1>
        <p>Create a new meeting or join one that already exists.</p>
        <div className="form-actions">
          <button type="button" className="primary-button" onClick={handleCreate}>
            Create a new meeting
          </button>
          <div className="section-divider">
            <span>or join using a meeting ID</span>
          </div>
          <form onSubmit={handleJoin}>
            <div className="input-group">
              <label htmlFor="meetingId">Meeting ID</label>
              <input
                id="meetingId"
                type="text"
                placeholder="Enter meeting ID"
                value={meetingId}
                onChange={(event) => setMeetingId(event.target.value)}
                required
              />
            </div>
            <div className="input-group">
              <label htmlFor="password">Passcode</label>
              <input
                id="password"
                type="text"
                placeholder="Passcode"
                value={password}
                onChange={(event) => setPassword(event.target.value)}
                required
              />
            </div>
            <button type="submit" className="secondary-button">
              Join meeting
            </button>
          </form>
        </div>
        {feedback && <div className="success-message">{feedback}</div>}
        {error && <div className="error-message">{error}</div>}
      </div>
    </div>
  );
}
