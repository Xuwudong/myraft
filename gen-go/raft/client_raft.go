// Code generated by Thrift Compiler (0.16.0). DO NOT EDIT.

package raft

import (
	"bytes"
	"context"
	"fmt"
	"time"
	thrift "github.com/apache/thrift/lib/go/thrift"

)

// (needed to ensure safety because of naive import list construction.)
var _ = thrift.ZERO
var _ = fmt.Printf
var _ = context.Background
var _ = time.Now
var _ = bytes.Equal

// Attributes:
//  - ID
//  - Command
type DoCommandReq struct {
  ID string `thrift:"id,1" db:"id" json:"id"`
  Command *Command `thrift:"command,2" db:"command" json:"command"`
}

func NewDoCommandReq() *DoCommandReq {
  return &DoCommandReq{}
}


func (p *DoCommandReq) GetID() string {
  return p.ID
}
var DoCommandReq_Command_DEFAULT *Command
func (p *DoCommandReq) GetCommand() *Command {
  if !p.IsSetCommand() {
    return DoCommandReq_Command_DEFAULT
  }
return p.Command
}
func (p *DoCommandReq) IsSetCommand() bool {
  return p.Command != nil
}

func (p *DoCommandReq) Read(ctx context.Context, iprot thrift.TProtocol) error {
  if _, err := iprot.ReadStructBegin(ctx); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T read error: ", p), err)
  }


  for {
    _, fieldTypeId, fieldId, err := iprot.ReadFieldBegin(ctx)
    if err != nil {
      return thrift.PrependError(fmt.Sprintf("%T field %d read error: ", p, fieldId), err)
    }
    if fieldTypeId == thrift.STOP { break; }
    switch fieldId {
    case 1:
      if fieldTypeId == thrift.STRING {
        if err := p.ReadField1(ctx, iprot); err != nil {
          return err
        }
      } else {
        if err := iprot.Skip(ctx, fieldTypeId); err != nil {
          return err
        }
      }
    case 2:
      if fieldTypeId == thrift.STRUCT {
        if err := p.ReadField2(ctx, iprot); err != nil {
          return err
        }
      } else {
        if err := iprot.Skip(ctx, fieldTypeId); err != nil {
          return err
        }
      }
    default:
      if err := iprot.Skip(ctx, fieldTypeId); err != nil {
        return err
      }
    }
    if err := iprot.ReadFieldEnd(ctx); err != nil {
      return err
    }
  }
  if err := iprot.ReadStructEnd(ctx); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T read struct end error: ", p), err)
  }
  return nil
}

func (p *DoCommandReq)  ReadField1(ctx context.Context, iprot thrift.TProtocol) error {
  if v, err := iprot.ReadString(ctx); err != nil {
  return thrift.PrependError("error reading field 1: ", err)
} else {
  p.ID = v
}
  return nil
}

func (p *DoCommandReq)  ReadField2(ctx context.Context, iprot thrift.TProtocol) error {
  p.Command = &Command{}
  if err := p.Command.Read(ctx, iprot); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T error reading struct: ", p.Command), err)
  }
  return nil
}

func (p *DoCommandReq) Write(ctx context.Context, oprot thrift.TProtocol) error {
  if err := oprot.WriteStructBegin(ctx, "DoCommandReq"); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write struct begin error: ", p), err) }
  if p != nil {
    if err := p.writeField1(ctx, oprot); err != nil { return err }
    if err := p.writeField2(ctx, oprot); err != nil { return err }
  }
  if err := oprot.WriteFieldStop(ctx); err != nil {
    return thrift.PrependError("write field stop error: ", err) }
  if err := oprot.WriteStructEnd(ctx); err != nil {
    return thrift.PrependError("write struct stop error: ", err) }
  return nil
}

func (p *DoCommandReq) writeField1(ctx context.Context, oprot thrift.TProtocol) (err error) {
  if err := oprot.WriteFieldBegin(ctx, "id", thrift.STRING, 1); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field begin error 1:id: ", p), err) }
  if err := oprot.WriteString(ctx, string(p.ID)); err != nil {
  return thrift.PrependError(fmt.Sprintf("%T.id (1) field write error: ", p), err) }
  if err := oprot.WriteFieldEnd(ctx); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field end error 1:id: ", p), err) }
  return err
}

func (p *DoCommandReq) writeField2(ctx context.Context, oprot thrift.TProtocol) (err error) {
  if err := oprot.WriteFieldBegin(ctx, "command", thrift.STRUCT, 2); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field begin error 2:command: ", p), err) }
  if err := p.Command.Write(ctx, oprot); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T error writing struct: ", p.Command), err)
  }
  if err := oprot.WriteFieldEnd(ctx); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field end error 2:command: ", p), err) }
  return err
}

func (p *DoCommandReq) Equals(other *DoCommandReq) bool {
  if p == other {
    return true
  } else if p == nil || other == nil {
    return false
  }
  if p.ID != other.ID { return false }
  if !p.Command.Equals(other.Command) { return false }
  return true
}

func (p *DoCommandReq) String() string {
  if p == nil {
    return "<nil>"
  }
  return fmt.Sprintf("DoCommandReq(%+v)", *p)
}

// Attributes:
//  - Succuess
//  - Value
//  - Leader
type DoCommandResp struct {
  Succuess bool `thrift:"succuess,1" db:"succuess" json:"succuess"`
  Value int64 `thrift:"value,2" db:"value" json:"value"`
  Leader *string `thrift:"leader,3" db:"leader" json:"leader,omitempty"`
}

func NewDoCommandResp() *DoCommandResp {
  return &DoCommandResp{}
}


func (p *DoCommandResp) GetSuccuess() bool {
  return p.Succuess
}

func (p *DoCommandResp) GetValue() int64 {
  return p.Value
}
var DoCommandResp_Leader_DEFAULT string
func (p *DoCommandResp) GetLeader() string {
  if !p.IsSetLeader() {
    return DoCommandResp_Leader_DEFAULT
  }
return *p.Leader
}
func (p *DoCommandResp) IsSetLeader() bool {
  return p.Leader != nil
}

func (p *DoCommandResp) Read(ctx context.Context, iprot thrift.TProtocol) error {
  if _, err := iprot.ReadStructBegin(ctx); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T read error: ", p), err)
  }


  for {
    _, fieldTypeId, fieldId, err := iprot.ReadFieldBegin(ctx)
    if err != nil {
      return thrift.PrependError(fmt.Sprintf("%T field %d read error: ", p, fieldId), err)
    }
    if fieldTypeId == thrift.STOP { break; }
    switch fieldId {
    case 1:
      if fieldTypeId == thrift.BOOL {
        if err := p.ReadField1(ctx, iprot); err != nil {
          return err
        }
      } else {
        if err := iprot.Skip(ctx, fieldTypeId); err != nil {
          return err
        }
      }
    case 2:
      if fieldTypeId == thrift.I64 {
        if err := p.ReadField2(ctx, iprot); err != nil {
          return err
        }
      } else {
        if err := iprot.Skip(ctx, fieldTypeId); err != nil {
          return err
        }
      }
    case 3:
      if fieldTypeId == thrift.STRING {
        if err := p.ReadField3(ctx, iprot); err != nil {
          return err
        }
      } else {
        if err := iprot.Skip(ctx, fieldTypeId); err != nil {
          return err
        }
      }
    default:
      if err := iprot.Skip(ctx, fieldTypeId); err != nil {
        return err
      }
    }
    if err := iprot.ReadFieldEnd(ctx); err != nil {
      return err
    }
  }
  if err := iprot.ReadStructEnd(ctx); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T read struct end error: ", p), err)
  }
  return nil
}

func (p *DoCommandResp)  ReadField1(ctx context.Context, iprot thrift.TProtocol) error {
  if v, err := iprot.ReadBool(ctx); err != nil {
  return thrift.PrependError("error reading field 1: ", err)
} else {
  p.Succuess = v
}
  return nil
}

func (p *DoCommandResp)  ReadField2(ctx context.Context, iprot thrift.TProtocol) error {
  if v, err := iprot.ReadI64(ctx); err != nil {
  return thrift.PrependError("error reading field 2: ", err)
} else {
  p.Value = v
}
  return nil
}

func (p *DoCommandResp)  ReadField3(ctx context.Context, iprot thrift.TProtocol) error {
  if v, err := iprot.ReadString(ctx); err != nil {
  return thrift.PrependError("error reading field 3: ", err)
} else {
  p.Leader = &v
}
  return nil
}

func (p *DoCommandResp) Write(ctx context.Context, oprot thrift.TProtocol) error {
  if err := oprot.WriteStructBegin(ctx, "DoCommandResp"); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write struct begin error: ", p), err) }
  if p != nil {
    if err := p.writeField1(ctx, oprot); err != nil { return err }
    if err := p.writeField2(ctx, oprot); err != nil { return err }
    if err := p.writeField3(ctx, oprot); err != nil { return err }
  }
  if err := oprot.WriteFieldStop(ctx); err != nil {
    return thrift.PrependError("write field stop error: ", err) }
  if err := oprot.WriteStructEnd(ctx); err != nil {
    return thrift.PrependError("write struct stop error: ", err) }
  return nil
}

func (p *DoCommandResp) writeField1(ctx context.Context, oprot thrift.TProtocol) (err error) {
  if err := oprot.WriteFieldBegin(ctx, "succuess", thrift.BOOL, 1); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field begin error 1:succuess: ", p), err) }
  if err := oprot.WriteBool(ctx, bool(p.Succuess)); err != nil {
  return thrift.PrependError(fmt.Sprintf("%T.succuess (1) field write error: ", p), err) }
  if err := oprot.WriteFieldEnd(ctx); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field end error 1:succuess: ", p), err) }
  return err
}

func (p *DoCommandResp) writeField2(ctx context.Context, oprot thrift.TProtocol) (err error) {
  if err := oprot.WriteFieldBegin(ctx, "value", thrift.I64, 2); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field begin error 2:value: ", p), err) }
  if err := oprot.WriteI64(ctx, int64(p.Value)); err != nil {
  return thrift.PrependError(fmt.Sprintf("%T.value (2) field write error: ", p), err) }
  if err := oprot.WriteFieldEnd(ctx); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field end error 2:value: ", p), err) }
  return err
}

func (p *DoCommandResp) writeField3(ctx context.Context, oprot thrift.TProtocol) (err error) {
  if p.IsSetLeader() {
    if err := oprot.WriteFieldBegin(ctx, "leader", thrift.STRING, 3); err != nil {
      return thrift.PrependError(fmt.Sprintf("%T write field begin error 3:leader: ", p), err) }
    if err := oprot.WriteString(ctx, string(*p.Leader)); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T.leader (3) field write error: ", p), err) }
    if err := oprot.WriteFieldEnd(ctx); err != nil {
      return thrift.PrependError(fmt.Sprintf("%T write field end error 3:leader: ", p), err) }
  }
  return err
}

func (p *DoCommandResp) Equals(other *DoCommandResp) bool {
  if p == other {
    return true
  } else if p == nil || other == nil {
    return false
  }
  if p.Succuess != other.Succuess { return false }
  if p.Value != other.Value { return false }
  if p.Leader != other.Leader {
    if p.Leader == nil || other.Leader == nil {
      return false
    }
    if (*p.Leader) != (*other.Leader) { return false }
  }
  return true
}

func (p *DoCommandResp) String() string {
  if p == nil {
    return "<nil>"
  }
  return fmt.Sprintf("DoCommandResp(%+v)", *p)
}

type ClientRaftServer interface {
  // Parameters:
  //  - Req
  DoCommand(ctx context.Context, req *DoCommandReq) (_r *DoCommandResp, _err error)
}

type ClientRaftServerClient struct {
  c thrift.TClient
  meta thrift.ResponseMeta
}

func NewClientRaftServerClientFactory(t thrift.TTransport, f thrift.TProtocolFactory) *ClientRaftServerClient {
  return &ClientRaftServerClient{
    c: thrift.NewTStandardClient(f.GetProtocol(t), f.GetProtocol(t)),
  }
}

func NewClientRaftServerClientProtocol(t thrift.TTransport, iprot thrift.TProtocol, oprot thrift.TProtocol) *ClientRaftServerClient {
  return &ClientRaftServerClient{
    c: thrift.NewTStandardClient(iprot, oprot),
  }
}

func NewClientRaftServerClient(c thrift.TClient) *ClientRaftServerClient {
  return &ClientRaftServerClient{
    c: c,
  }
}

func (p *ClientRaftServerClient) Client_() thrift.TClient {
  return p.c
}

func (p *ClientRaftServerClient) LastResponseMeta_() thrift.ResponseMeta {
  return p.meta
}

func (p *ClientRaftServerClient) SetLastResponseMeta_(meta thrift.ResponseMeta) {
  p.meta = meta
}

// Parameters:
//  - Req
func (p *ClientRaftServerClient) DoCommand(ctx context.Context, req *DoCommandReq) (_r *DoCommandResp, _err error) {
  var _args0 ClientRaftServerDoCommandArgs
  _args0.Req = req
  var _result2 ClientRaftServerDoCommandResult
  var _meta1 thrift.ResponseMeta
  _meta1, _err = p.Client_().Call(ctx, "DoCommand", &_args0, &_result2)
  p.SetLastResponseMeta_(_meta1)
  if _err != nil {
    return
  }
  if _ret3 := _result2.GetSuccess(); _ret3 != nil {
    return _ret3, nil
  }
  return nil, thrift.NewTApplicationException(thrift.MISSING_RESULT, "DoCommand failed: unknown result")
}

type ClientRaftServerProcessor struct {
  processorMap map[string]thrift.TProcessorFunction
  handler ClientRaftServer
}

func (p *ClientRaftServerProcessor) AddToProcessorMap(key string, processor thrift.TProcessorFunction) {
  p.processorMap[key] = processor
}

func (p *ClientRaftServerProcessor) GetProcessorFunction(key string) (processor thrift.TProcessorFunction, ok bool) {
  processor, ok = p.processorMap[key]
  return processor, ok
}

func (p *ClientRaftServerProcessor) ProcessorMap() map[string]thrift.TProcessorFunction {
  return p.processorMap
}

func NewClientRaftServerProcessor(handler ClientRaftServer) *ClientRaftServerProcessor {

  self4 := &ClientRaftServerProcessor{handler:handler, processorMap:make(map[string]thrift.TProcessorFunction)}
  self4.processorMap["DoCommand"] = &clientRaftServerProcessorDoCommand{handler:handler}
return self4
}

func (p *ClientRaftServerProcessor) Process(ctx context.Context, iprot, oprot thrift.TProtocol) (success bool, err thrift.TException) {
  name, _, seqId, err2 := iprot.ReadMessageBegin(ctx)
  if err2 != nil { return false, thrift.WrapTException(err2) }
  if processor, ok := p.GetProcessorFunction(name); ok {
    return processor.Process(ctx, seqId, iprot, oprot)
  }
  iprot.Skip(ctx, thrift.STRUCT)
  iprot.ReadMessageEnd(ctx)
  x5 := thrift.NewTApplicationException(thrift.UNKNOWN_METHOD, "Unknown function " + name)
  oprot.WriteMessageBegin(ctx, name, thrift.EXCEPTION, seqId)
  x5.Write(ctx, oprot)
  oprot.WriteMessageEnd(ctx)
  oprot.Flush(ctx)
  return false, x5

}

type clientRaftServerProcessorDoCommand struct {
  handler ClientRaftServer
}

func (p *clientRaftServerProcessorDoCommand) Process(ctx context.Context, seqId int32, iprot, oprot thrift.TProtocol) (success bool, err thrift.TException) {
  args := ClientRaftServerDoCommandArgs{}
  var err2 error
  if err2 = args.Read(ctx, iprot); err2 != nil {
    iprot.ReadMessageEnd(ctx)
    x := thrift.NewTApplicationException(thrift.PROTOCOL_ERROR, err2.Error())
    oprot.WriteMessageBegin(ctx, "DoCommand", thrift.EXCEPTION, seqId)
    x.Write(ctx, oprot)
    oprot.WriteMessageEnd(ctx)
    oprot.Flush(ctx)
    return false, thrift.WrapTException(err2)
  }
  iprot.ReadMessageEnd(ctx)

  tickerCancel := func() {}
  // Start a goroutine to do server side connectivity check.
  if thrift.ServerConnectivityCheckInterval > 0 {
    var cancel context.CancelFunc
    ctx, cancel = context.WithCancel(ctx)
    defer cancel()
    var tickerCtx context.Context
    tickerCtx, tickerCancel = context.WithCancel(context.Background())
    defer tickerCancel()
    go func(ctx context.Context, cancel context.CancelFunc) {
      ticker := time.NewTicker(thrift.ServerConnectivityCheckInterval)
      defer ticker.Stop()
      for {
        select {
        case <-ctx.Done():
          return
        case <-ticker.C:
          if !iprot.Transport().IsOpen() {
            cancel()
            return
          }
        }
      }
    }(tickerCtx, cancel)
  }

  result := ClientRaftServerDoCommandResult{}
  var retval *DoCommandResp
  if retval, err2 = p.handler.DoCommand(ctx, args.Req); err2 != nil {
    tickerCancel()
    if err2 == thrift.ErrAbandonRequest {
      return false, thrift.WrapTException(err2)
    }
    x := thrift.NewTApplicationException(thrift.INTERNAL_ERROR, "Internal error processing DoCommand: " + err2.Error())
    oprot.WriteMessageBegin(ctx, "DoCommand", thrift.EXCEPTION, seqId)
    x.Write(ctx, oprot)
    oprot.WriteMessageEnd(ctx)
    oprot.Flush(ctx)
    return true, thrift.WrapTException(err2)
  } else {
    result.Success = retval
  }
  tickerCancel()
  if err2 = oprot.WriteMessageBegin(ctx, "DoCommand", thrift.REPLY, seqId); err2 != nil {
    err = thrift.WrapTException(err2)
  }
  if err2 = result.Write(ctx, oprot); err == nil && err2 != nil {
    err = thrift.WrapTException(err2)
  }
  if err2 = oprot.WriteMessageEnd(ctx); err == nil && err2 != nil {
    err = thrift.WrapTException(err2)
  }
  if err2 = oprot.Flush(ctx); err == nil && err2 != nil {
    err = thrift.WrapTException(err2)
  }
  if err != nil {
    return
  }
  return true, err
}


// HELPER FUNCTIONS AND STRUCTURES

// Attributes:
//  - Req
type ClientRaftServerDoCommandArgs struct {
  Req *DoCommandReq `thrift:"req,1" db:"req" json:"req"`
}

func NewClientRaftServerDoCommandArgs() *ClientRaftServerDoCommandArgs {
  return &ClientRaftServerDoCommandArgs{}
}

var ClientRaftServerDoCommandArgs_Req_DEFAULT *DoCommandReq
func (p *ClientRaftServerDoCommandArgs) GetReq() *DoCommandReq {
  if !p.IsSetReq() {
    return ClientRaftServerDoCommandArgs_Req_DEFAULT
  }
return p.Req
}
func (p *ClientRaftServerDoCommandArgs) IsSetReq() bool {
  return p.Req != nil
}

func (p *ClientRaftServerDoCommandArgs) Read(ctx context.Context, iprot thrift.TProtocol) error {
  if _, err := iprot.ReadStructBegin(ctx); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T read error: ", p), err)
  }


  for {
    _, fieldTypeId, fieldId, err := iprot.ReadFieldBegin(ctx)
    if err != nil {
      return thrift.PrependError(fmt.Sprintf("%T field %d read error: ", p, fieldId), err)
    }
    if fieldTypeId == thrift.STOP { break; }
    switch fieldId {
    case 1:
      if fieldTypeId == thrift.STRUCT {
        if err := p.ReadField1(ctx, iprot); err != nil {
          return err
        }
      } else {
        if err := iprot.Skip(ctx, fieldTypeId); err != nil {
          return err
        }
      }
    default:
      if err := iprot.Skip(ctx, fieldTypeId); err != nil {
        return err
      }
    }
    if err := iprot.ReadFieldEnd(ctx); err != nil {
      return err
    }
  }
  if err := iprot.ReadStructEnd(ctx); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T read struct end error: ", p), err)
  }
  return nil
}

func (p *ClientRaftServerDoCommandArgs)  ReadField1(ctx context.Context, iprot thrift.TProtocol) error {
  p.Req = &DoCommandReq{}
  if err := p.Req.Read(ctx, iprot); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T error reading struct: ", p.Req), err)
  }
  return nil
}

func (p *ClientRaftServerDoCommandArgs) Write(ctx context.Context, oprot thrift.TProtocol) error {
  if err := oprot.WriteStructBegin(ctx, "DoCommand_args"); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write struct begin error: ", p), err) }
  if p != nil {
    if err := p.writeField1(ctx, oprot); err != nil { return err }
  }
  if err := oprot.WriteFieldStop(ctx); err != nil {
    return thrift.PrependError("write field stop error: ", err) }
  if err := oprot.WriteStructEnd(ctx); err != nil {
    return thrift.PrependError("write struct stop error: ", err) }
  return nil
}

func (p *ClientRaftServerDoCommandArgs) writeField1(ctx context.Context, oprot thrift.TProtocol) (err error) {
  if err := oprot.WriteFieldBegin(ctx, "req", thrift.STRUCT, 1); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field begin error 1:req: ", p), err) }
  if err := p.Req.Write(ctx, oprot); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T error writing struct: ", p.Req), err)
  }
  if err := oprot.WriteFieldEnd(ctx); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write field end error 1:req: ", p), err) }
  return err
}

func (p *ClientRaftServerDoCommandArgs) String() string {
  if p == nil {
    return "<nil>"
  }
  return fmt.Sprintf("ClientRaftServerDoCommandArgs(%+v)", *p)
}

// Attributes:
//  - Success
type ClientRaftServerDoCommandResult struct {
  Success *DoCommandResp `thrift:"success,0" db:"success" json:"success,omitempty"`
}

func NewClientRaftServerDoCommandResult() *ClientRaftServerDoCommandResult {
  return &ClientRaftServerDoCommandResult{}
}

var ClientRaftServerDoCommandResult_Success_DEFAULT *DoCommandResp
func (p *ClientRaftServerDoCommandResult) GetSuccess() *DoCommandResp {
  if !p.IsSetSuccess() {
    return ClientRaftServerDoCommandResult_Success_DEFAULT
  }
return p.Success
}
func (p *ClientRaftServerDoCommandResult) IsSetSuccess() bool {
  return p.Success != nil
}

func (p *ClientRaftServerDoCommandResult) Read(ctx context.Context, iprot thrift.TProtocol) error {
  if _, err := iprot.ReadStructBegin(ctx); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T read error: ", p), err)
  }


  for {
    _, fieldTypeId, fieldId, err := iprot.ReadFieldBegin(ctx)
    if err != nil {
      return thrift.PrependError(fmt.Sprintf("%T field %d read error: ", p, fieldId), err)
    }
    if fieldTypeId == thrift.STOP { break; }
    switch fieldId {
    case 0:
      if fieldTypeId == thrift.STRUCT {
        if err := p.ReadField0(ctx, iprot); err != nil {
          return err
        }
      } else {
        if err := iprot.Skip(ctx, fieldTypeId); err != nil {
          return err
        }
      }
    default:
      if err := iprot.Skip(ctx, fieldTypeId); err != nil {
        return err
      }
    }
    if err := iprot.ReadFieldEnd(ctx); err != nil {
      return err
    }
  }
  if err := iprot.ReadStructEnd(ctx); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T read struct end error: ", p), err)
  }
  return nil
}

func (p *ClientRaftServerDoCommandResult)  ReadField0(ctx context.Context, iprot thrift.TProtocol) error {
  p.Success = &DoCommandResp{}
  if err := p.Success.Read(ctx, iprot); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T error reading struct: ", p.Success), err)
  }
  return nil
}

func (p *ClientRaftServerDoCommandResult) Write(ctx context.Context, oprot thrift.TProtocol) error {
  if err := oprot.WriteStructBegin(ctx, "DoCommand_result"); err != nil {
    return thrift.PrependError(fmt.Sprintf("%T write struct begin error: ", p), err) }
  if p != nil {
    if err := p.writeField0(ctx, oprot); err != nil { return err }
  }
  if err := oprot.WriteFieldStop(ctx); err != nil {
    return thrift.PrependError("write field stop error: ", err) }
  if err := oprot.WriteStructEnd(ctx); err != nil {
    return thrift.PrependError("write struct stop error: ", err) }
  return nil
}

func (p *ClientRaftServerDoCommandResult) writeField0(ctx context.Context, oprot thrift.TProtocol) (err error) {
  if p.IsSetSuccess() {
    if err := oprot.WriteFieldBegin(ctx, "success", thrift.STRUCT, 0); err != nil {
      return thrift.PrependError(fmt.Sprintf("%T write field begin error 0:success: ", p), err) }
    if err := p.Success.Write(ctx, oprot); err != nil {
      return thrift.PrependError(fmt.Sprintf("%T error writing struct: ", p.Success), err)
    }
    if err := oprot.WriteFieldEnd(ctx); err != nil {
      return thrift.PrependError(fmt.Sprintf("%T write field end error 0:success: ", p), err) }
  }
  return err
}

func (p *ClientRaftServerDoCommandResult) String() string {
  if p == nil {
    return "<nil>"
  }
  return fmt.Sprintf("ClientRaftServerDoCommandResult(%+v)", *p)
}


