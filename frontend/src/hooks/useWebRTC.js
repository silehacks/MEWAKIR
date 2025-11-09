import { useEffect, useMemo, useRef, useState } from 'react';

const ICE_SERVERS = [
  { urls: 'stun:stun.l.google.com:19302' },
  { urls: 'stun:stun1.l.google.com:19302' },
];

function buildSignalingUrl(meetingId, name) {
  const base = window.location.origin.startsWith('https')
    ? window.location.origin.replace('https', 'wss')
    : window.location.origin.replace('http', 'ws');
  return `${base}/ws/signaling?meetingId=${encodeURIComponent(meetingId)}&name=${encodeURIComponent(name)}`;
}

export function useWebRTC({ meetingId, displayName }) {
  const [localStream, setLocalStream] = useState(null);
  const [remoteStreams, setRemoteStreams] = useState([]);
  const [audioEnabled, setAudioEnabled] = useState(true);
  const [videoEnabled, setVideoEnabled] = useState(true);
  const [error, setError] = useState(null);

  const pcRef = useRef(null);
  const wsRef = useRef(null);
  const remoteStreamMap = useRef(new Map());

  const signalingUrl = useMemo(() => buildSignalingUrl(meetingId, displayName), [meetingId, displayName]);

  useEffect(() => {
    let isMounted = true;

    async function init() {
      try {
        const media = await navigator.mediaDevices.getUserMedia({ audio: true, video: true });
        if (!isMounted) return;
        setLocalStream(media);

        const pc = new RTCPeerConnection({ iceServers: ICE_SERVERS });
        pcRef.current = pc;

        media.getTracks().forEach((track) => pc.addTrack(track, media));

        pc.onicecandidate = (event) => {
          if (event.candidate) {
            sendSignal({ type: 'candidate', candidate: event.candidate });
          }
        };

        pc.ontrack = (event) => {
          const [stream] = event.streams;
          if (!stream) return;
          const id = stream.id;
          remoteStreamMap.current.set(id, stream);
          setRemoteStreams(Array.from(remoteStreamMap.current.values()));
        };

        pc.onconnectionstatechange = () => {
          if (pc.connectionState === 'failed') {
            setError('Connection failed. Please try rejoining the meeting.');
          }
        };

        const ws = new WebSocket(signalingUrl);
        wsRef.current = ws;

        ws.onopen = () => {
          sendSignal({ type: 'join', name: displayName });
        };

        ws.onmessage = async (event) => {
          try {
            const message = JSON.parse(event.data);
            if (!message || message.from === displayName) return;

            switch (message.type) {
              case 'join':
                await createAndSendOffer();
                break;
              case 'offer':
                await handleOffer(message.sdp);
                break;
              case 'answer':
                await handleAnswer(message.sdp);
                break;
              case 'candidate':
                if (message.candidate) {
                  await pc.addIceCandidate(message.candidate);
                }
                break;
              default:
                break;
            }
          } catch (err) {
            console.error('Failed to process signaling message', err);
          }
        };

        ws.onclose = () => {
          remoteStreamMap.current.clear();
          setRemoteStreams([]);
        };

        ws.onerror = () => {
          setError('Signaling server unreachable.');
        };
      } catch (err) {
        console.error(err);
        setError(err.message || 'Unable to access camera or microphone.');
      }
    }

    init();

    return () => {
      isMounted = false;
      localStream?.getTracks().forEach((track) => track.stop());
      if (pcRef.current) {
        pcRef.current.close();
        pcRef.current = null;
      }
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [signalingUrl]);

  async function createAndSendOffer() {
    const pc = pcRef.current;
    if (!pc) return;
    const offer = await pc.createOffer();
    await pc.setLocalDescription(offer);
    sendSignal({ type: 'offer', sdp: offer });
  }

  async function handleOffer(offer) {
    const pc = pcRef.current;
    if (!pc) return;
    await pc.setRemoteDescription(new RTCSessionDescription(offer));
    const answer = await pc.createAnswer();
    await pc.setLocalDescription(answer);
    sendSignal({ type: 'answer', sdp: answer });
  }

  async function handleAnswer(answer) {
    const pc = pcRef.current;
    if (!pc) return;
    await pc.setRemoteDescription(new RTCSessionDescription(answer));
  }

  function sendSignal(message) {
    const ws = wsRef.current;
    if (!ws || ws.readyState !== WebSocket.OPEN) return;
    ws.send(JSON.stringify({ ...message, from: displayName }));
  }

  function toggleAudio() {
    if (!localStream) return;
    const enabled = !audioEnabled;
    localStream.getAudioTracks().forEach((track) => {
      track.enabled = enabled;
    });
    setAudioEnabled(enabled);
  }

  function toggleVideo() {
    if (!localStream) return;
    const enabled = !videoEnabled;
    localStream.getVideoTracks().forEach((track) => {
      track.enabled = enabled;
    });
    setVideoEnabled(enabled);
  }

  return {
    localStream,
    remoteStreams,
    audioEnabled,
    videoEnabled,
    toggleAudio,
    toggleVideo,
    error,
  };
}
