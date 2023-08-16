package goevents

import goerror "github.com/start-codex/errors"

const ErrPayloadNil = goerror.Error("interface{} is nil")
const ErrPayloadTypeAssertion = goerror.Error("interface{} is not of type %T")
const ErrNoExitsTopic = goerror.Error("Topic is not exists")
