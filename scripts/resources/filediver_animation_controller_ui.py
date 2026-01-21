import bpy
try:
    import bpy.stub_internal
except:
    pass
from typing import Set, Optional, List, Tuple

import math
import bl_math
import random

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

class filediver_curve(bpy.types.PropertyGroup):
    data_path: bpy.props.StringProperty()

class filediver_frame(bpy.types.PropertyGroup):
    index: bpy.props.FloatProperty()

class filediver_keyframe_group(bpy.types.PropertyGroup):
    start: bpy.props.FloatProperty()
    end: bpy.props.FloatProperty()
    event: bpy.props.StringProperty()
    group_id: bpy.props.FloatProperty()
    action: bpy.props.PointerProperty(type=bpy.types.Action)
    fcurves_path: bpy.props.StringProperty()
    curves: bpy.props.CollectionProperty(type=filediver_curve)

    def resolve(self) -> List[bpy.types.Keyframe]:
        action: bpy.types.Action = self.action
        fcurves: bpy.types.ActionChannelbagFCurves = action.path_resolve(self.fcurves_path)
        to_return: List[bpy.types.Keyframe] = []
        for curve in self.curves:
            curve: filediver_curve
            fcurve = fcurves.find(curve.data_path)
            for keyframe in fcurve.keyframe_points:
                if keyframe.back == self.group_id:
                    to_return.append(keyframe)
        return to_return

    def remove(self) -> None:
        action: bpy.types.Action = self.action
        fcurves: bpy.types.ActionChannelbagFCurves = action.path_resolve(self.fcurves_path)
        to_remove = 0
        for curve in self.curves:
            curve: filediver_curve
            fcurve = fcurves.find(curve.data_path)
            for keyframe in fcurve.keyframe_points:
                if keyframe.back == self.group_id:
                    to_remove += 1
            while to_remove > 0:
                fcurve = fcurves.find(curve.data_path)
                for keyframe in fcurve.keyframe_points:
                    if keyframe.back == self.group_id:
                        fcurve.keyframe_points.remove(keyframe, fast=True)
                        to_remove -= 1
                        break


FILEDIVER_ANIMATION_VARIABLES = ["state", "next_state", "state_transition", "start_frame", "phase_frame"]
FILEDIVER_MAX_FLOAT_INT = 9007199254740992

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

    def filter_items(self, context, data, property) -> Tuple[List[int], List[int]]:
        filter_flags: List[int] = []
        new_item_order: List[int] = []

        events = getattr(data, property)

        helpers = bpy.types.UI_UL_list

        if self.filter_name:
            filter_flags = helpers.filter_items_by_name(self.filter_name, self.bitflag_filter_item, events, "name", reverse=self.use_filter_sort_reverse)

        if not filter_flags:
            filter_flags = [self.bitflag_filter_item] * len(events)
        elif self.use_filter_invert:
            for i in range(len(filter_flags)):
                filter_flags[i] = filter_flags[i] ^ self.bitflag_filter_item

        return filter_flags, new_item_order

class SCENE_UL_filediver_animation_events(bpy.types.UIList):
    def draw_item(self, context, layout: bpy.types.UILayout, data, item, icon, active_data, active_propname, index):
        layout.prop(item, "event", text="", emboss=False)

    def filter_items(self, context, data, property) -> Tuple[List[int], List[int]]:
        filter_flags: List[int] = []
        new_item_order: List[int] = []

        events = getattr(data, property)

        helpers = bpy.types.UI_UL_list

        if self.filter_name:
            filter_flags = helpers.filter_items_by_name(self.filter_name, self.bitflag_filter_item, events, "event", reverse=self.use_filter_sort_reverse)

        if not filter_flags:
            filter_flags = [self.bitflag_filter_item] * len(events)
        elif self.use_filter_invert:
            for i in range(len(filter_flags)):
                filter_flags[i] = filter_flags[i] ^ self.bitflag_filter_item

        return filter_flags, new_item_order

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

        def insert_keyframe(points: bpy.types.FCurveKeyframePoints, frame: float, value: float, interpolation: "bpy.stub_internal.rna_enums.BeztripleInterpolationModeItems", group_id: float):
            keyframe = points.insert(frame, value, options={'FAST'})
            keyframe.interpolation = interpolation
            keyframe.back = group_id

        while len(end_events) > 0:
            current_frame, current_event = end_events.pop(0)
            context.scene.frame_set(current_frame)
            group_id = float(random.randrange(FILEDIVER_MAX_FLOAT_INT))
            for layer in sorted_layers:
                layer: bpy.types.Object
                layer_strip: bpy.types.ActionStrip = layer.animation_data.action.layers[0].strips[0]
                layer_keyframe_strip: bpy.types.ActionKeyframeStrip = None
                if layer_strip.type == "KEYFRAME":
                    layer_keyframe_strip = layer_strip
                layer_channelbag = layer_keyframe_strip.channelbags[0]
                layer_fcurves = layer_channelbag.fcurves

                state_curve = layer_fcurves.find("state")
                if state_curve is None:
                    state_curve = layer_fcurves.new("state")

                next_state_curve = layer_fcurves.find("next_state")
                if next_state_curve is None:
                    next_state_curve = layer_fcurves.new("next_state")

                transition_curve = layer_fcurves.find("state_transition")
                if transition_curve is None:
                    transition_curve = layer_fcurves.new("state_transition")

                start_curve = None
                phase_curve = None
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
                insert_keyframe(transition_curve.keyframe_points, start_frame, 0, 'LINEAR', group_id)
                if start_frame != end_frame:
                    insert_keyframe(next_state_curve.keyframe_points, start_frame, transition.state_index, 'CONSTANT', group_id)
                    insert_keyframe(state_curve.keyframe_points, end_frame+1, transition.state_index, 'CONSTANT', group_id)
                    insert_keyframe(transition_curve.keyframe_points, end_frame, 1, 'LINEAR', group_id)
                    insert_keyframe(transition_curve.keyframe_points, end_frame+1, 0, 'LINEAR', group_id)
                else:
                    insert_keyframe(state_curve.keyframe_points, start_frame, transition.state_index, 'CONSTANT', group_id)

                insert_keyframe(next_state_curve.keyframe_points, end_frame+1, -1, 'CONSTANT', group_id)

                if not new_state.loop:
                    start_curve = layer_fcurves.find("start_frame")
                    if start_curve is None:
                        start_curve = layer_fcurves.new("start_frame")
                    insert_keyframe(start_curve.keyframe_points, start_frame, float(current_frame), 'CONSTANT', group_id)
                else:
                    phase_curve = layer_fcurves.find("phase_frame")
                    if phase_curve is None:
                        phase_curve = layer_fcurves.new("phase_frame")
                    insert_keyframe(phase_curve.keyframe_points, start_frame, float(current_frame), 'CONSTANT', group_id)

                if new_state.emit_end_event != "":
                    end_events.append((new_state.get_next_end_frame(end_frame+1), new_state.emit_end_event))

                key_group: filediver_keyframe_group = state_machine.filediver_keyframe_groups.add()
                key_group.start = start_frame
                key_group.end = end_frame
                for curve in [state_curve, next_state_curve, transition_curve, start_curve, phase_curve]:
                    curve: Optional[bpy.types.FCurve]
                    if curve is None:
                        continue
                    fd_curve: filediver_curve = key_group.curves.add()
                    fd_curve.data_path = curve.data_path
                key_group.event = current_event
                key_group.group_id = group_id
                key_group.fcurves_path = layer_fcurves.path_from_id()
                key_group.action = layer.animation_data.action

                layer.filediver_applying_transition = False

        context.scene.frame_set(old_frame)

        return {'FINISHED'}

class OBJECT_OT_fd_remove_transition(bpy.types.Operator):
    bl_idname = "object.fd_remove_transition"
    bl_label = "Remove"

    bl_options = {'REGISTER', 'UNDO'}

    state_machine_name: bpy.props.StringProperty()
    group_id: bpy.props.FloatProperty()

    def execute(self, context: bpy.types.Context) -> Set["bpy.stub_internal.rna_enums.OperatorReturnItems"]:
        if self.state_machine_name is None:
            self.report({'DEBUG'}, "cancelled")
            return {'CANCELLED'}

        state_machine = context.scene.objects[self.state_machine_name]
        to_remove = []
        for index, group in enumerate(state_machine.filediver_keyframe_groups):
            group: filediver_keyframe_group
            if group.group_id != self.group_id:
                continue
            group.remove()
            to_remove.append(index)

        for index in reversed(to_remove):
            state_machine.filediver_keyframe_groups.remove(index)

        return {'FINISHED'}

class OBJECT_OT_fd_remove_all_transitions(bpy.types.Operator):
    bl_idname = "object.fd_remove_all_transitions"
    bl_label = "Remove All"

    bl_options = {'REGISTER', 'UNDO'}

    state_machine_name: bpy.props.StringProperty()

    def execute(self, context: bpy.types.Context) -> Set["bpy.stub_internal.rna_enums.OperatorReturnItems"]:
        if self.state_machine_name is None:
            self.report({'DEBUG'}, "cancelled")
            return {'CANCELLED'}

        state_machine = context.scene.objects[self.state_machine_name]
        for group in state_machine.filediver_keyframe_groups:
            group.remove()
        state_machine.filediver_keyframe_groups.clear()

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

        header, body = layout.panel(state_machine.name + "_event_list", default_closed=False)
        header.label(text="Sent Events")
        if body is not None and len(state_machine.filediver_keyframe_groups) > 0:
            ids = set()
            for group in state_machine.filediver_keyframe_groups:
                group: filediver_keyframe_group
                if group.group_id in ids:
                    continue
                row = body.row()
                row.label(text=group.event + f" @ frame {int(group.start)}")
                delete_props = row.operator(OBJECT_OT_fd_remove_transition.bl_idname)
                delete_props.state_machine_name = state_machine.name
                delete_props.group_id = group.group_id
                ids.add(group.group_id)
            delete_props = body.operator(OBJECT_OT_fd_remove_all_transitions.bl_idname)
            delete_props.state_machine_name = state_machine.name
        elif body is not None:
            body.label(text="None")


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
    bpy.utils.register_class(filediver_curve)
    bpy.utils.register_class(filediver_frame)
    bpy.utils.register_class(filediver_keyframe_group)
    bpy.utils.register_class(OBJECT_OT_fd_transition_animation)
    bpy.utils.register_class(OBJECT_OT_fd_remove_transition)
    bpy.utils.register_class(OBJECT_OT_fd_remove_all_transitions)
    bpy.utils.register_class(SCENE_UL_filediver_animation_states)
    bpy.utils.register_class(SCENE_UL_filediver_animation_events)
    bpy.utils.register_class(SCENE_PT_filediver_animation_controller)
    bpy.types.Object.filediver_layer_states = bpy.props.CollectionProperty(type=filediver_animation_state)
    bpy.types.Object.filediver_keyframe_groups = bpy.props.CollectionProperty(type=filediver_keyframe_group)
    bpy.types.Object.state = bpy.props.IntProperty(update=update_layer_state)
    bpy.types.Object.next_state = bpy.props.IntProperty(default=-1)
    bpy.types.Object.state_transition = bpy.props.FloatProperty(min=0, max=1)
    bpy.types.Object.start_frame = bpy.props.FloatProperty()
    bpy.types.Object.phase_frame = bpy.props.FloatProperty()
    bpy.types.Object.filediver_applying_transition = bpy.props.BoolProperty()
    bpy.types.Object.filediver_active_transitions = bpy.props.CollectionProperty(type=filediver_event_name)
    bpy.types.Object.filediver_transition_index = bpy.props.IntProperty(update=update_controller_events)

def unregister():
    bpy.utils.unregister_class(filediver_keyframe_group)
    bpy.utils.unregister_class(filediver_frame)
    bpy.utils.unregister_class(filediver_curve)
    bpy.utils.unregister_class(filediver_event_name)
    bpy.utils.unregister_class(filediver_animation_state)
    bpy.utils.unregister_class(filediver_state_transition)
    bpy.utils.unregister_class(filediver_animation_variable)
    bpy.utils.unregister_class(filediver_animation)
    bpy.utils.unregister_class(OBJECT_OT_fd_remove_all_transitions)
    bpy.utils.unregister_class(OBJECT_OT_fd_remove_transition)
    bpy.utils.unregister_class(OBJECT_OT_fd_transition_animation)
    bpy.utils.unregister_class(SCENE_PT_filediver_animation_controller)
    bpy.utils.unregister_class(SCENE_UL_filediver_animation_events)
    bpy.utils.unregister_class(SCENE_UL_filediver_animation_states)

if __name__ == "__main__":
    register()