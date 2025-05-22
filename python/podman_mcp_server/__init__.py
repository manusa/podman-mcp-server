"""
Model Context Protocol (MCP) server for container runtimes (Podman and Docker)
"""
from .podman_mcp_server import main

__all__ = ['main']

