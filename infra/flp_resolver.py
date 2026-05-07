import sys
import json
import enum

# Monkeypatch EventEnum for Python 3.12+ compatibility
try:
    import pyflp._events
    if not hasattr(pyflp._events.EventEnum, '__members__') or len(pyflp._events.EventEnum) == 0:
        pyflp._events.EventEnum = enum.IntEnum("EventEnum", names=())
except ImportError:
    pass

import pyflp

def resolve_flp(file_path):
    try:
        project = pyflp.parse(file_path)
        plugins = []
        
        # Iterate over channels to find plugins
        for channel in project.channels:
            if hasattr(channel, 'plugin') and channel.plugin is not None:
                plugins.append({
                    "name": channel.name,
                    "plugin_name": channel.plugin.name if hasattr(channel.plugin, 'name') else "Unknown",
                    "type": "Channel"
                })
        
        # Also check mixer tracks for effects
        for track in project.mixer:
            for slot in track.slots:
                if slot.plugin is not None:
                    plugins.append({
                        "name": f"Mixer {track.index} Slot {slot.index}",
                        "plugin_name": slot.plugin.name if hasattr(slot.plugin, 'name') else "Unknown",
                        "type": "Effect"
                    })

        return {"plugins": plugins, "error": None}
    except Exception as e:
        return {"plugins": [], "error": str(e)}

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print(json.dumps({"plugins": [], "error": "No file path provided"}))
        sys.exit(1)
    
    file_path = sys.argv[1]
    result = resolve_flp(file_path)
    print(json.dumps(result))
