import bpy
try:
    import bpy.stub_internal
except:
    pass
from typing import Set, Optional

import math
import bl_math

link_types = [
    ("LinkType_Immediate", "Immediate", "", 0),
    ("LinkType_WaitEnd", "WaitEnd", "", 1),
    ("LinkType_SyncBeatClosestImmediate", "SyncBeatClosestImmediate", "", 2),
    ("LinkType_WaitUntilBeat", "WaitUntilBeat", "", 3),
    ("LinkType_SyncPercentageImmediate", "SyncPercentageImmediate", "", 4),
    ("LinkType_SyncInversePercentageImmediate", "SyncInversePercentageImmediate", "", 5),
]

class filediver_state_transition(bpy.types.PropertyGroup):
    event: bpy.props.StringProperty()
    state_index: bpy.props.IntProperty()
    blend_time: bpy.props.FloatProperty()
    link_type: bpy.props.EnumProperty(items=link_types)
    beat: bpy.props.StringProperty()

class filediver_event_name(bpy.types.PropertyGroup):
    event: bpy.props.StringProperty()

class filediver_animation(bpy.types.PropertyGroup):
    name: bpy.props.StringProperty()

class filediver_animation_variable(bpy.types.PropertyGroup):
    name: bpy.props.StringProperty()
    obj: bpy.props.PointerProperty(type=bpy.types.Object)

FILEDIVER_ANIMATION_VARIABLES = ["state", "next_state", "state_transition", "start_frame", "phase_frame"]

class filediver_animation_state(bpy.types.PropertyGroup):
    name: bpy.props.StringProperty()
    type: bpy.props.StringProperty()
    emit_end_event: bpy.props.StringProperty()
    animations: bpy.props.CollectionProperty(type=filediver_animation)
    loop: bpy.props.BoolProperty()
    transitions: bpy.props.CollectionProperty(type=filediver_state_transition)
    active_index: bpy.props.IntProperty()
    variables: bpy.props.CollectionProperty(type=filediver_animation_variable)
    frequency_expr: bpy.props.StringProperty()
    animation_length: bpy.props.FloatProperty()

    def get_transition(self, event: str) -> Optional[filediver_state_transition]:
        for transition in self.transitions:
            if event == transition.event:
                return transition
        return None
    
    def get_next_end_frame(self, frame: int) -> int:
        if self.frequency_expr == "" and self.animation_length == 0.0:
            return frame
        eval_globals = {
            'fps': bpy.data.scenes[0].render.fps,
            'frame': frame,
            'floor': math.floor,
            'clamp': bl_math.clamp
        }
        for variable in self.variables:
            variable: filediver_animation_variable
            if variable.name in eval_globals:
                continue
            eval_globals[variable.name] = variable.obj.get(variable.name)

        if self.animation_length > 0 and eval_globals["start_frame"] + self.animation_length > frame:
            return int(math.ceil(eval_globals["start_frame"] + self.animation_length))
        elif self.animation_length > 0 and eval_globals["start_frame"] + self.animation_length <= frame:
            return frame
        frequency = eval(self.frequency_expr, eval_globals)
        period = float(eval_globals["fps"]) / frequency
        return int(math.ceil(eval_globals["phase_frame"] + period * (((frame - eval_globals["phase_frame"]) // period)+1)))


class SCENE_UL_filediver_animation_states(bpy.types.UIList):
    def draw_item(self, context, layout, data, item, icon, active_data, active_propname, index):
        layout.prop(item, "name", text="", emboss=False)

class SCENE_UL_filediver_animation_events(bpy.types.UIList):
    def draw_item(self, context, layout: bpy.types.UILayout, data, item, icon, active_data, active_propname, index):
        layout.prop(item, "event", text="", emboss=False)

class OBJECT_OT_fd_transition_animation(bpy.types.Operator):
    bl_idname = "object.fd_transition_animation"
    bl_label = "Send Animation Event"

    bl_options = {'REGISTER', 'UNDO'}

    state_machine_name: bpy.props.StringProperty()
    event: bpy.props.StringProperty()

    def execute(self, context: bpy.types.Context) -> Set["bpy.stub_internal.rna_enums.OperatorReturnItems"]:
        if self.state_machine_name is None:
            self.report({'DEBUG'}, "cancelled")
            return {'CANCELLED'}
        print(f"Sending event {self.event}")
        #self.report({'DEBUG'}, self.state_machine_name)
        state_machine = context.scene.objects[self.state_machine_name]
        sorted_layers = sorted(state_machine.children, key=lambda x: int(x.name.split()[-1]))
        end_events = [(context.scene.frame_current, self.event)]
        old_frame = context.scene.frame_current
        while len(end_events) > 0:
            current_frame, current_event = end_events.pop(0)
            context.scene.frame_set(current_frame)
            for layer in sorted_layers:
                layer: bpy.types.Object
                if layer.state >= len(layer.filediver_layer_states):
                    print("layer.state >= len(states)")
                    continue
                state: filediver_animation_state = layer.filediver_layer_states[layer.state]
                transition = state.get_transition(current_event)
                if transition is None:
                    print("transition is None")
                    continue
                new_state: filediver_animation_state = layer.filediver_layer_states[transition.state_index]
                old_state = layer.state
                layer.filediver_applying_transition = True
                if transition.link_type in ("LinkType_Immediate", "LinkType_SyncPercentageImmediate"):
                    start_frame = current_frame
                    end_frame = current_frame+int(context.scene.render.fps*transition.blend_time)
                elif transition.link_type == "LinkType_WaitEnd":
                    end_frame = state.get_next_end_frame(current_frame)
                    start_frame = end_frame - int(context.scene.render.fps*transition.blend_time)
                else:
                    self.report({'DEBUG'}, transition.link_type)

                # apply transition keyframes
                print(f"Adding keyframes to layer {layer.name} for event {current_event}")
                layer.state = transition.state_index
                layer.state_transition = 0
                layer.keyframe_insert(data_path="state_transition", frame=start_frame)
                if start_frame != end_frame:
                    layer.next_state = transition.state_index
                    layer.keyframe_insert(data_path="next_state", frame=start_frame)
                    layer.keyframe_insert(data_path="state", frame=end_frame+1)
                    layer.state_transition = 1
                    layer.keyframe_insert(data_path="state_transition", frame=end_frame)
                    layer.state_transition = 0
                    layer.keyframe_insert(data_path="state_transition", frame=end_frame+1)
                else:
                    layer.keyframe_insert(data_path="state", frame=start_frame)
                layer.next_state = -1
                layer.keyframe_insert(data_path="next_state", frame=end_frame+1)

                if not new_state.loop:
                    layer.start_frame = float(current_frame)
                    layer.keyframe_insert(data_path="start_frame", frame=start_frame)
                else:
                    layer.phase_frame = float(current_frame)
                    layer.keyframe_insert(data_path="phase_frame", frame=start_frame)

                if new_state.emit_end_event != "":
                    end_events.append((new_state.get_next_end_frame(end_frame+1), new_state.emit_end_event))

                strips: bpy.types.ActionKeyframeStrip = layer.animation_data.action.layers[0].strips[0]
                fcurves = strips.channelbag(layer.animation_data.action_slot).fcurves
                for fcurve in fcurves:
                    interpolation = 'CONSTANT'
                    if "state_transition" in fcurve.data_path:
                        interpolation = 'LINEAR'
                    for keyframe in fcurve.keyframe_points:
                        keyframe.interpolation = interpolation
                layer.state = old_state
                layer.filediver_applying_transition = False

        context.scene.frame_set(old_frame)

        return {'FINISHED'}

def not_a_controller(row: bpy.types.UILayout):
    row.label(text="No animation controller selected")

class SCENE_PT_filediver_animation_controller(bpy.types.Panel):
    bl_label = "Filediver Animation Controller Info"
    bl_space_type = 'VIEW_3D'
    bl_region_type = 'UI'
    bl_category = "Filediver"

    def draw(self, context):
        layout = self.layout

        if not context.active_object:
            return not_a_controller(layout)
        if not (context.active_object.data is None or type(context.active_object.data) == bpy.types.Armature):
            return not_a_controller(layout)
        
        state_machine = None
        if ".state_machine" in context.active_object.name:
            state_machine = context.active_object
        else:
            for child in context.active_object.children:
                if ".state_machine" in child.name:
                    state_machine = child
                    break
        if state_machine is None:
            return not_a_controller(layout)

        row = layout.row()
        row.label(text=state_machine.name)

        layout.label(text="Animation Events")
        layout.template_list("SCENE_UL_filediver_animation_events", "", state_machine, "filediver_active_transitions", state_machine, "filediver_transition_index")
        transition_props = layout.operator(OBJECT_OT_fd_transition_animation.bl_idname)
        transition_props.state_machine_name = state_machine.name
        if state_machine.filediver_transition_index < len(state_machine.filediver_active_transitions):
            transition_props.event = state_machine.filediver_active_transitions[state_machine.filediver_transition_index].event
        else:
            transition_props.event = ""

        sorted_layers = sorted(state_machine.children, key=lambda x: int(x.name.split()[-1]))
        for layer in sorted_layers:
            header, body = layout.panel(layer.name, default_closed=True)
            header.label(text=layer.name)
            if body is None:
                continue
            body.label(text="States")
            body.template_list("SCENE_UL_filediver_animation_states", f"list_{layer.name}", layer, "filediver_layer_states", layer, "state")
            body.label(text="Variables")
            for variable in layer.filediver_layer_states[layer.state].variables:
                if variable.name in FILEDIVER_ANIMATION_VARIABLES:
                    # filediver animation variables are not accessed as custom properties
                    body.prop(variable.obj, variable.name)
                    continue
                body.prop(variable.obj, f'["{variable.name}"]')


def update_layer_state(self: bpy.types.Object, context):
    if "layer" not in self.name:
        return
    if len(self.filediver_layer_states) > 0:
        if self.state >= len(self.filediver_layer_states):
            self.state = len(self.filediver_layer_states) - 1
        if self.state < 0:
            self.state = 0
    if not self.parent or "state_machine" not in self.parent.name:
        return
    if self.filediver_applying_transition:
        return
    used_events: Set[str] = set()
    for layer in self.parent.children:
        if "layer" not in layer.name or len(layer.filediver_layer_states) <= layer.state:
            continue
        for transition in layer.filediver_layer_states[layer.state].transitions:
            transition: filediver_state_transition
            used_events.add(transition.event)
    self.parent.filediver_active_transitions.clear()
    for event in sorted(list(used_events)):
        transition: filediver_event_name = self.parent.filediver_active_transitions.add()
        transition.event = event

def update_controller_events(self: bpy.types.Object, context):
    if "state_machine" not in self.name:
        return
    if self.filediver_transition_index < len(self.filediver_active_transitions):
        current_selection = self.filediver_active_transitions[self.filediver_transition_index].event
    else:
        current_selection = ""
    used_events: Set[str] = set()
    for layer in self.children:
        if "layer" not in layer.name or len(layer.filediver_layer_states) <= layer.state:
            continue
        for transition in layer.filediver_layer_states[layer.state].transitions:
            transition: filediver_state_transition
            used_events.add(transition.event)
    self.filediver_active_transitions.clear()
    found_idx = -1
    for idx, event in enumerate(sorted(list(used_events))):
        if current_selection == event:
            found_idx = idx
        transition: filediver_event_name = self.filediver_active_transitions.add()
        transition.event = event

    # if found_idx != -1:
    #     self.filediver_transition_index = found_idx
    # else:
    #     self.filediver_transition_index = max(0, min(self.filediver_transition_index, len(self.filediver_active_transitions)-1))
    #self.filediver_transition_index = max(0, min(self.filediver_transition_index, len(self.filediver_active_transitions)-1))

def register():
    bpy.utils.register_class(filediver_animation)
    bpy.utils.register_class(filediver_animation_variable)
    bpy.utils.register_class(filediver_state_transition)
    bpy.utils.register_class(filediver_animation_state)
    bpy.utils.register_class(filediver_event_name)
    bpy.utils.register_class(OBJECT_OT_fd_transition_animation)
    bpy.utils.register_class(SCENE_UL_filediver_animation_states)
    bpy.utils.register_class(SCENE_UL_filediver_animation_events)
    bpy.utils.register_class(SCENE_PT_filediver_animation_controller)
    bpy.types.Object.filediver_layer_states = bpy.props.CollectionProperty(type=filediver_animation_state)
    bpy.types.Object.state = bpy.props.IntProperty(update=update_layer_state)
    bpy.types.Object.next_state = bpy.props.IntProperty(default=-1)
    bpy.types.Object.state_transition = bpy.props.FloatProperty(min=0, max=1)
    bpy.types.Object.start_frame = bpy.props.FloatProperty()
    bpy.types.Object.phase_frame = bpy.props.FloatProperty()
    bpy.types.Object.filediver_applying_transition = bpy.props.BoolProperty()
    bpy.types.Object.filediver_active_transitions = bpy.props.CollectionProperty(type=filediver_event_name)
    bpy.types.Object.filediver_transition_index = bpy.props.IntProperty(update=update_controller_events)

def unregister():
    bpy.utils.unregister_class(filediver_event_name)
    bpy.utils.unregister_class(filediver_animation_state)
    bpy.utils.unregister_class(filediver_state_transition)
    bpy.utils.unregister_class(filediver_animation_variable)
    bpy.utils.unregister_class(filediver_animation)
    bpy.utils.unregister_class(OBJECT_OT_fd_transition_animation)
    bpy.utils.unregister_class(SCENE_PT_filediver_animation_controller)
    bpy.utils.unregister_class(SCENE_UL_filediver_animation_events)
    bpy.utils.unregister_class(SCENE_UL_filediver_animation_states)

if __name__ == "__main__":
    register()