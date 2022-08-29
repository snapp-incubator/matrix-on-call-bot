CREATE TABLE IF NOT EXISTS follow_ups (
    id INT NOT NULL AUTO_INCREMENT,
    shift_id INT NOT NULL,
    sender TEXT NOT NULL,
    initiator TEXT NOT NULL,
    description TEXT,
    done BOOLEAN NOT NULL,
    category TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    FOREIGN KEY (shift_id) REFERENCES shifts(id)
);