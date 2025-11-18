import bpy

# class SCENE_UL_filediver_animation_events(bpy.types.UIList):
#     def draw_item(self, context, layout, data, item, icon, active_data, active_propname, index):
#         layout.prop(item, "name", text="", emboss=False)

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
        for layer in state_machine.children:
            row = layout.row()
            row.label(text=layer.name)
            #row.template_list("SCENE_UL_filediver_animation_events", "", layer, "states", )

def register():
    bpy.utils.register_class(SCENE_PT_filediver_animation_controller)

def unregister():
    bpy.utils.unregister_class(SCENE_PT_filediver_animation_controller)

if __name__ == "__main__":
    register()