import bpy
from bpy.app.handlers import persistent

def state_influence(state_idx, mask_influence, current_state, next_state, transition_pct):
    return mask_influence*float(current_state==state_idx)*(1-transition_pct)+float(next_state==state_idx)*transition_pct

def start_frame(state_idx, current_state, next_state, start_frame, next_start_frame):
    if current_state == state_idx:
        return start_frame
    if next_state == state_idx:
        return next_start_frame
    return 0

@persistent
def load_handler(dummy):
    bpy.app.driver_namespace['infl'] = state_influence
    bpy.app.driver_namespace['start'] = start_frame

def register():
    load_handler(None)
    bpy.app.handlers.load_post.append(load_handler)

def unregister():
    bpy.app.handlers.load_post.remove(load_handler)

if __name__ == "__main__":
    register()