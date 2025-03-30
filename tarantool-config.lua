box.cfg{
    listen = 3301
  }

-- Создаём пользователя для подключения
box.schema.user.create('taracan', {password='tarantool_pwd', if_not_exists=true})
-- Даём все-все права
box.schema.user.grant('taracan', 'super', nil, nil, {if_not_exists=true})
-- Чуть настраиваем сериализатор в iproto, чтобы не ругался на несериализуемые типы
require('msgpack').cfg{encode_invalid_as_nil = true}

-- Создаём space для хранения опросов
if not box.space.polls then
    box.schema.create_space('polls', {
        format = {
            {name = 'id', type = 'string'},
            {name = 'question', type = 'string'},
            {name = 'options', type = 'array'},
            {name = 'votes', type = 'map'},
            {name = 'creator', type = 'string'},
            {name = 'isactive',   type = 'boolean'},
        },
        if_not_exists = true
    })
    box.space.polls:create_index('primary', {
        parts = {'id'},
        if_not_exists = true
    })
end