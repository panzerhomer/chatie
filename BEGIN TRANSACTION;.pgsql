BEGIN TRANSACTION;

INSERT INTO employees(email) VALUES ('test1@mail.com');
INSERT INTO users(email, password) VALUES ('test1@mail.com', 'password');

COMMIT TRANSACTION;

INSERT INTO employees(email) VALUES ('test2@mail.com');
INSERT INTO users(email, password) VALUES ('test2@mail.com', 'password');

INSERT INTO employees(email) VALUES ('test3@mail.com');
INSERT INTO users(email, password) VALUES ('test3@mail.com', 'password');

select * from users;
select * from employees;

select * from users join employees on users.email = employees.email;


INSERT INTO chats(owner_id, label, info, is_private) VALUES (1, 'info', 'chat1', false);
INSERT INTO chat_members(chat_id, user_id, user_role, status) VALUES (1, 5, 'user', 'unbanned');

select * from chats;

select * from chat_members;

select * from chats join chat_members 
on chats.chat_id = chat_members.chat_id where chats.chat_id = 1

INSERT INTO messages(text, from_id, chat_id) VALUES ('hello from test1', 1, 1)
INSERT INTO messages(text, from_id, chat_id) VALUES ('hello from test1', 2, 1)


select * from messages