from abc import ABC, abstractmethod
from typing import Dict, Optional
import bpy.types

class FilediverMaterialLoaderInterface(ABC):
    """Interface for material loaders"""

    @abstractmethod
    def load_material(self, resource_path: str) -> None:
        """Load the material from the template file"""

    @abstractmethod
    def add_material(self, config: dict, textures: Dict[str, bpy.types.Image]) -> bpy.types.Material:
        """Configure a copy of this loader's material and return it"""

    @abstractmethod
    def get_material(self, config: dict, index: int) -> Optional[bpy.types.Material]:
        """Search for an existing copy of this configuration and return it if found"""

    @abstractmethod
    def preprocess_config(self, data: bpy.types.BlendData, gltf: dict, materialTextures: Dict[int, Dict[str, bpy.types.Image]], config: dict) -> dict:
        """Apply any required preprocessing to the config using bpy.data and the gltf file"""

    @classmethod
    @abstractmethod
    def can_configure(cls, config: dict) -> bool:
        """Returns true if this class can configure a material with the given config"""

    @classmethod
    @abstractmethod
    def key(cls) -> str:
        """Returns a string identifying the type of the material"""