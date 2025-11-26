import bpy

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

class filediver_animation_state(bpy.types.PropertyGroup):
    name: bpy.props.StringProperty()
    transitions: bpy.props.CollectionProperty(type=filediver_state_transition)

class SCENE_UL_filediver_animation_events(bpy.types.UIList):
    def draw_item(self, context, layout, data, item, icon, active_data, active_propname, index):
        layout.prop(item, "name", text="", emboss=False)

def not_a_controller(row: bpy.types.UILayout):
    row.label(text="No animation controller selected")

class SCENE_PT_filediver_animation_controller(bpy.types.Panel):
    bl_label = "Filediver Animation Controller Info"
    bl_space_type = 'VIEW_3D'
    bl_region_type = 'UI'
    bl_category = "Stingray"

    def draw(self, context):
        layout = self.layout
        row = layout.row()

        if not context.active_object:
            return not_a_controller(row)
        if not (context.active_object.data is None or type(context.active_object.data) == bpy.types.Armature):
            return not_a_controller(row)
        
        state_machine = None
        if ".state_machine" in context.active_object.name:
            state_machine = context.active_object
        else:
            for child in context.active_object.children:
                if ".state_machine" in child.name:
                    state_machine = child
                    break
        if state_machine is None:
            return not_a_controller(row)
        
        row.label(text=state_machine.name)
        sorted_layers = sorted(state_machine.children, key=lambda x: int(x.name.split()[-1]))
        for layer in sorted_layers:
            row = layout.row()
            row.label(text=layer.name)
            col = row.column()
            col.template_list("SCENE_UL_filediver_animation_events", "", layer, "filediver_layer_states", layer, "state")

def register():
    bpy.utils.register_class(filediver_state_transition)
    bpy.utils.register_class(filediver_animation_state)
    bpy.utils.register_class(SCENE_UL_filediver_animation_events)
    bpy.utils.register_class(SCENE_PT_filediver_animation_controller)
    bpy.types.Object.filediver_layer_states = bpy.props.CollectionProperty(type=filediver_animation_state)
    bpy.types.Object.state = bpy.props.IntProperty()

def unregister():
    bpy.utils.unregister_class(filediver_animation_state)
    bpy.utils.unregister_class(filediver_state_transition)
    bpy.utils.unregister_class(SCENE_PT_filediver_animation_controller)
    bpy.utils.unregister_class(SCENE_UL_filediver_animation_events)

if __name__ == "__main__":
    register()