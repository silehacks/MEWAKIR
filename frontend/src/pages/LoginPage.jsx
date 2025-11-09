import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { api } from '../hooks/useApi';
import { useAppContext } from '../context/AppContext';

export function LoginPage() {
  const [name, setName] = useState('');
  const [error, setError] = useState(null);
  const navigate = useNavigate();
  const { setDisplayName } = useAppContext();

  async function handleSubmit(event) {
    event.preventDefault();
    setError(null);
    const trimmed = name.trim();
    if (!trimmed) {
      setError('Please enter your name to continue.');
      return;
    }
    try {
      await api.login(trimmed);
      setDisplayName(trimmed);
      navigate('/home');
    } catch (err) {
      setError(err.message || 'Unable to sign in.');
    }
  }

  return (
    <div className="main-layout">
      <form className="card" onSubmit={handleSubmit}>
        <h1>Welcome to Mewakir</h1>
        <p>Enter your name to join meetings with the people who matter most.</p>
        <div className="input-group">
          <label htmlFor="displayName">Your name</label>
          <input
            id="displayName"
            type="text"
            placeholder="Jane Doe"
            value={name}
            onChange={(event) => setName(event.target.value)}
          />
        </div>
        {error && <div className="error-message">{error}</div>}
        <button type="submit" className="primary-button">
          Continue
        </button>
      </form>
    </div>
  );
}
