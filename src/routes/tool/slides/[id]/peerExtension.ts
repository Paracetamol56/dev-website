import { collab, getSyncedVersion, receiveUpdates, sendableUpdates, type Update } from "@codemirror/collab";
import { ViewPlugin, type ViewUpdate } from "@codemirror/view";
import { Text, type Extension } from '@codemirror/state';
import { type EditorView } from "codemirror";

function generateRandomID() {
  return (
    Math.random().toString(36).substring(2, 15) + Math.random().toString(36).substring(2, 15)
  );
}

async function wsRequest(ws: WebSocket, message: string): Promise<any> {
  return new Promise((resolve, _) => {
    const channel = new MessageChannel();
    channel.port1.onmessage = (event) => {
      resolve(JSON.parse(event.data));
    }
    ws.send(message, [channel.port2]);
  });
}

async function pushUpdates(ws: WebSocket, version: number, fullUpdates: readonly Update[]) {
  const updates = fullUpdates.map((u) => ({
    clientID: u.clientID,
    changes: u.changes.toJSON()
  }));
  ws.send(JSON.stringify({
    type: "pushUpdates",
    version,
    updates,
  }));
}

async function getDocument(ws: WebSocket) {
  return wsRequest(ws, JSON.stringify({
    type: "getDocument"
  })).then((data) => ({
    version: data.version,
    doc: Text.of(data.doc.split("\n"))
  }));
}

export default function peerExtension(startVersion: number, ws: WebSocket): Extension {
  const plugin = ViewPlugin.fromClass(
    class {
      private view: EditorView;
      private ws: WebSocket = ws;
      private pushing = false;

      constructor(private view: EditorView) {
        this.ws.onmessage = (event) => {
          const data = JSON.parse(event.data);
          console.log("Received data:", data);
          if (data.type === "pushUpdates") {
            this.view.dispatch(receiveUpdates(this.view.state, data.updates as readonly Update[]));
          }
        }
      }

      update(update: ViewUpdate) {
        if (update.docChanged) this.push();
      }

      async push() {
        const updates = sendableUpdates(this.view.state);
        console.log("Pushing updates:", updates);
        if (this.pushing || !updates.length) return;
        this.pushing = true;
        const version = getSyncedVersion(this.view.state);
        await pushUpdates(this.ws, version, updates);
        this.pushing = false;
        if (sendableUpdates(this.view.state).length > 0) {
          setTimeout(() => this.push(), 100);
        }
      }
    }
  );

  return [collab({ clientID: generateRandomID(), startVersion }), plugin];
}

