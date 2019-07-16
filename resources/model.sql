CREATE TABLE Users
(
    id         VARCHAR(10) NOT NULL,
    channel_id VARCHAR(10) NOT NULL,
    mobile     VARCHAR(10) NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE Spaces
(
    id           INT         NOT NULL AUTO_INCREMENT,
    number_space VARCHAR(5)  NOT NULL,
    available    TINYINT     NOT NULL,
    block_id     INT         NULL,
    id_user      VARCHAR(10) NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (id_user) REFERENCES Users (id)
);

CREATE TABLE Vehicles
(
    id      VARCHAR(10) NOT NULL,
    type    VARCHAR(10) NOT NULL,
    brand   VARCHAR(20) NOT NULL,
    color   VARCHAR(10) NOT NULL,
    id_user VARCHAR(10) NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (id_user) REFERENCES Users (id)
);