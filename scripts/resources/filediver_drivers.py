import bpy
from bpy.app.handlers import persistent

def state_influence(state_idx, mask_influence, current_state, next_state, transition_pct):
    return mask_influence*float(current_state==state_idx)*(1-transition_pct)+float(next_state==state_idx)*transition_pct

@persistent
def load_handler(dummy):
    bpy.app.driver_namespace['infl'] = state_influence

def register():
    load_handler(None)
    bpy.app.handlers.load_post.append(load_handler)

def unregister():
    bpy.app.handlers.load_post.remove(load_handler)

if __name__ == "__main__":
    register()