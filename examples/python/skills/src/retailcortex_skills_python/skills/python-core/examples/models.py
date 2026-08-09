# Copyright 2026 Ryan McGuinness
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

from __future__ import annotations
from enum import Enum
from typing import List, Optional
from uuid import UUID, uuid4
from pydantic import BaseModel, EmailStr
from sqlmodel import Field, Relationship, SQLModel

class CustomerType(str, Enum):
    INDIVIDUAL = "INDIVIDUAL"
    ENTERPRISE = "ENTERPRISE"

class CustomerBase(SQLModel):
    name: str = Field(index=True)
    email: EmailStr = Field(unique=True, index=True)
    customer_type: CustomerType = Field(default=CustomerType.INDIVIDUAL)
    is_active: bool = Field(default=True)

class Customer(CustomerBase, table=True):
    __tablename__ = "customers"

    id: UUID = Field(default_factory=uuid4, primary_key=True, index=True)
    orders: List[Order] = Relationship(back_populates="customer")

class OrderBase(SQLModel):
    amount: float = Field(ge=0.0)
    currency: str = Field(default="USD", max_length=3)
    status: str = Field(default="PENDING")

class Order(OrderBase, table=True):
    __tablename__ = "orders"

    id: UUID = Field(default_factory=uuid4, primary_key=True)
    customer_id: UUID = Field(foreign_key="customers.id")
    customer: Optional[Customer] = Relationship(back_populates="orders")

# Pydantic schema for API response output
class CustomerRead(CustomerBase):
    id: UUID
    orders: List[OrderBase] = []
