ALTER TABLE certificates ADD COLUMN session_id VARCHAR(255);
CREATE INDEX idx_certificates_session_id ON certificates(session_id);