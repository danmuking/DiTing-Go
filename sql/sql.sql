-- auto-generated definition
create table contact
(
    id          bigint unsigned auto_increment comment 'id'
        primary key,
    uid         bigint                                   not null comment 'uid',
    room_id     bigint                                   not null comment '房间id',
    read_time   datetime(3) default CURRENT_TIMESTAMP(3) not null comment '阅读到的时间',
    active_time datetime(3)                              null comment '会话内消息最后更新的时间(只有普通会话需要维护，全员会话不需要维护)',
    last_msg_id bigint                                   null comment '会话最新消息id',
    create_time datetime(3) default CURRENT_TIMESTAMP(3) not null comment '创建时间',
    update_time datetime(3) default CURRENT_TIMESTAMP(3) not null on update CURRENT_TIMESTAMP(3) comment '修改时间',
    constraint uniq_uid_room_id
        unique (uid, room_id)
)
    comment '会话列表' collate = utf8mb4_unicode_ci;

create index idx_create_time
    on contact (create_time);

create index idx_room_id_read_time
    on contact (room_id, read_time);

create index idx_update_time
    on contact (update_time);

create index idx_user_id_active_time
    on contact (uid, active_time);

-- auto-generated definition
create table group_member
(
    id          bigint unsigned auto_increment comment 'id'
        primary key,
    group_id    bigint                                   not null comment '群主id',
    uid         bigint                                   not null comment '成员uid',
    role        int                                      not null comment '成员角色 1群主 2管理员 3普通成员',
    create_time datetime(3) default CURRENT_TIMESTAMP(3) not null comment '创建时间',
    update_time datetime(3) default CURRENT_TIMESTAMP(3) not null on update CURRENT_TIMESTAMP(3) comment '修改时间'
)
    comment '群成员表' collate = utf8mb4_unicode_ci;

create index idx_create_time
    on group_member (create_time);

create index idx_group_id_role
    on group_member (group_id, role);

create index idx_update_time
    on group_member (update_time);

-- auto-generated definition
create table message
(
    id           bigint unsigned auto_increment comment 'id'
        primary key,
    room_id      bigint                                   not null comment '会话表id',
    from_uid     bigint                                   not null comment '消息发送者uid',
    content      varchar(1024)                            null comment '消息内容',
    reply_msg_id bigint                                   null comment '回复的消息内容',
    status       int                                      not null comment '消息状态 0正常 1删除',
    gap_count    int                                      null comment '与回复的消息间隔多少条',
    type         int         default 1                    null comment '消息类型 1正常文本 2.撤回消息',
    extra        json                                     null comment '扩展信息',
    create_time  datetime(3) default CURRENT_TIMESTAMP(3) not null comment '创建时间',
    update_time  datetime(3) default CURRENT_TIMESTAMP(3) not null on update CURRENT_TIMESTAMP(3) comment '修改时间'
)
    comment '消息表' collate = utf8mb4_unicode_ci
                     row_format = DYNAMIC;

create index idx_create_time
    on message (create_time);

create index idx_from_uid
    on message (from_uid);

create index idx_room_id
    on message (room_id);

create index idx_update_time
    on message (update_time);

-- auto-generated definition
create table room
(
    id          bigint unsigned auto_increment comment 'id'
        primary key,
    type        int                                      not null comment '房间类型 1群聊 2单聊',
    hot_flag    int         default 0                    null comment '是否全员展示 0否 1是',
    active_time datetime(3) default CURRENT_TIMESTAMP(3) not null comment '群最后消息的更新时间（热点群不需要写扩散，只更新这里）',
    last_msg_id bigint                                   null comment '会话中的最后一条消息id',
    ext_json    json                                     null comment '额外信息（根据不同类型房间有不同存储的东西）',
    create_time datetime(3) default CURRENT_TIMESTAMP(3) not null comment '创建时间',
    update_time datetime(3) default CURRENT_TIMESTAMP(3) not null on update CURRENT_TIMESTAMP(3) comment '修改时间'
)
    comment '房间表' collate = utf8mb4_unicode_ci;

create index idx_create_time
    on room (create_time);

create index idx_update_time
    on room (update_time);

-- auto-generated definition
create table room_friend
(
    id          bigint unsigned auto_increment comment 'id'
        primary key,
    room_id     bigint                                   not null comment '房间id',
    uid1        bigint                                   not null comment 'uid1（更小的uid）',
    uid2        bigint                                   not null comment 'uid2（更大的uid）',
    room_key    varchar(64)                              not null comment '房间key由两个uid拼接，先做排序uid1_uid2',
    status      int                                      not null comment '房间状态 0正常 1禁用(删好友了禁用)',
    create_time datetime(3) default CURRENT_TIMESTAMP(3) not null comment '创建时间',
    update_time datetime(3) default CURRENT_TIMESTAMP(3) not null on update CURRENT_TIMESTAMP(3) comment '修改时间',
    constraint room_key
        unique (room_key)
)
    comment '单聊房间表' collate = utf8mb4_unicode_ci;

create index idx_create_time
    on room_friend (create_time);

create index idx_room_id
    on room_friend (room_id);

create index idx_update_time
    on room_friend (update_time);

-- auto-generated definition
create table room_group
(
    id            bigint unsigned auto_increment comment 'id'
        primary key,
    room_id       bigint                                   not null comment '房间id',
    name          varchar(16)                              not null comment '群名称',
    avatar        varchar(256)                             not null comment '群头像',
    ext_json      json                                     null comment '额外信息（根据不同类型房间有不同存储的东西）',
    delete_status int(1)      default 0                    not null comment '逻辑删除(0-正常,1-删除)',
    create_time   datetime(3) default CURRENT_TIMESTAMP(3) not null comment '创建时间',
    update_time   datetime(3) default CURRENT_TIMESTAMP(3) not null on update CURRENT_TIMESTAMP(3) comment '修改时间'
)
    comment '群聊房间表' collate = utf8mb4_unicode_ci;

create index idx_create_time
    on room_group (create_time);

create index idx_room_id
    on room_group (room_id);

create index idx_update_time
    on room_group (update_time);

-- auto-generated definition
create table user_friend
(
    id            bigint unsigned auto_increment comment 'id'
        primary key,
    uid           bigint                                   not null comment 'uid',
    friend_uid    bigint                                   not null comment '好友uid',
    delete_status int(1)      default 0                    not null comment '逻辑删除(0-正常,1-删除)',
    create_time   datetime(3) default CURRENT_TIMESTAMP(3) not null comment '创建时间',
    update_time   datetime(3) default CURRENT_TIMESTAMP(3) not null on update CURRENT_TIMESTAMP(3) comment '修改时间',
    constraint uid
        unique (uid, friend_uid)
)
    comment '用户联系人表' collate = utf8mb4_unicode_ci;

create index idx_create_time
    on user_friend (create_time);

create index idx_uid_friend_uid
    on user_friend (uid, friend_uid);

create index idx_update_time
    on user_friend (update_time);

-- auto-generated definition
create table user_apply
(
    id          bigint unsigned auto_increment comment 'id'
        primary key,
    uid         bigint                                   not null comment '申请人uid',
    type        int                                      not null comment '申请类型 1加好友',
    target_id   bigint                                   not null comment '接收人uid',
    msg         varchar(64)                              not null comment '申请信息',
    status      int                                      not null comment '申请状态 1待审批 2同意',
    read_status int                                      not null comment '阅读状态 1未读 2已读',
    create_time datetime(3) default CURRENT_TIMESTAMP(3) not null comment '创建时间',
    update_time datetime(3) default CURRENT_TIMESTAMP(3) not null on update CURRENT_TIMESTAMP(3) comment '修改时间',
    constraint uid
        unique (uid, target_id)
)
    comment '用户申请表' collate = utf8mb4_unicode_ci;

create index idx_create_time
    on user_apply (create_time);

create index idx_target_id
    on user_apply (target_id);

create index idx_target_id_read_status
    on user_apply (target_id, read_status);

create index idx_uid_target_id
    on user_apply (uid, target_id);

create index idx_update_time
    on user_apply (update_time);

