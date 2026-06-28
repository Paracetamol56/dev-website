
class OperationalTransformEditor {
  private documentId: string;
  private ws: any;
  private revision: number;
  private pendingOps: any[];

  constructor(documentId: string, ws: any) {
    this.documentId = documentId;
    this.ws = ws;
    this.revision = 0;
    this.pendingOps = [];
  }
  
  function createOperation() { }

function onLocalOp() { }

function onRemoteOp() { }
}
